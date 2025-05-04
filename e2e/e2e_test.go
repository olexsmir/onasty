package e2e_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/olexsmir/onasty/internal/config"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/jwtutil"
	"github.com/olexsmir/onasty/internal/logger"
	"github.com/olexsmir/onasty/internal/service/notesrv"
	"github.com/olexsmir/onasty/internal/service/usersrv"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/store/psql/sessionrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psql/vertokrepo"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
	"github.com/olexsmir/onasty/internal/store/rdb"
	"github.com/olexsmir/onasty/internal/store/rdb/notecache"
	"github.com/olexsmir/onasty/internal/store/rdb/usercache"
	httptransport "github.com/olexsmir/onasty/internal/transport/http"
	"github.com/olexsmir/onasty/internal/transport/http/ratelimit"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type (
	stopFunc     func()
	AppTestSuite struct {
		suite.Suite

		ctx     context.Context
		require *require.Assertions

		postgresDB   *psqlutil.DB
		stopPostgres stopFunc

		redisDB   *rdb.DB
		stopRedis stopFunc

		router       http.Handler
		hasher       hasher.Hasher
		jwtTokenizer jwtutil.JWTTokenizer
	}
	errorResponse struct {
		Message string `json:"message"`
	}
)

func TestAppSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// gin output is too verbose(and annoying) in tests
	gin.SetMode(gin.TestMode)

	suite.Run(t, new(AppTestSuite))
}

func (e *AppTestSuite) SetupSuite() {
	e.ctx = context.Background()
	e.require = e.Require()

	e.postgresDB, e.stopPostgres = e.prepPostgres()
	e.redisDB, e.stopRedis = e.prepRedis()

	e.initDeps()
}

func (e *AppTestSuite) TearDownSuite() {
	e.stopPostgres()
	e.stopRedis()
}

// initDeps initializes the dependencies for the app
// and sets up the router for tests
func (e *AppTestSuite) initDeps() {
	cfg := e.getConfig()

	logger, err := logger.NewCustomLogger(cfg.LogLevel, cfg.LogFormat, cfg.LogShowLine)
	e.require.NoError(err)

	slog.SetDefault(logger)

	e.hasher = hasher.NewSHA256Hasher(cfg.PasswordSalt)
	e.jwtTokenizer = jwtutil.NewJWTUtil(cfg.JwtSigningKey, time.Hour)

	sessionrepo := sessionrepo.New(e.postgresDB)
	vertokrepo := vertokrepo.New(e.postgresDB)

	oatuhProvider := newOauthProviderMock()

	userepo := userepo.New(e.postgresDB)
	usercache := usercache.New(e.redisDB, cfg.CacheUsersTTL)
	usersrv := usersrv.New(
		userepo,
		sessionrepo,
		vertokrepo,
		e.hasher,
		e.jwtTokenizer,
		newMailerMockService(),
		usercache,
		oatuhProvider,
		oatuhProvider,
		cfg.JwtRefreshTokenTTL,
		cfg.VerificationTokenTTL,
	)

	notecache := notecache.New(e.redisDB, cfg.CacheUsersTTL)
	noterepo := noterepo.New(e.postgresDB)
	notesrv := notesrv.New(noterepo, e.hasher, notecache)

	// for testing purposes, it's ok to have high values ig
	ratelimitCfg := ratelimit.Config{
		RPS:   1000,
		TTL:   time.Millisecond,
		Burst: 1000,
	}

	handler := httptransport.NewTransport(usersrv, notesrv, ratelimitCfg)
	e.router = handler.Handler()
}

func (e *AppTestSuite) prepPostgres() (*psqlutil.DB, stopFunc) {
	dbCredential := "testing"
	postgresContainer, err := tcpostgres.Run(e.ctx,
		"postgres:16-alpine",
		tcpostgres.WithUsername(dbCredential),
		tcpostgres.WithPassword(dbCredential),
		tcpostgres.WithDatabase(dbCredential),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")))
	e.require.NoError(err)

	stop := func() { e.require.NoError(postgresContainer.Terminate(e.ctx)) }

	// connect to the db
	host, err := postgresContainer.Host(e.ctx)
	e.require.NoError(err)

	port, err := postgresContainer.MappedPort(e.ctx, "5432/tcp")
	e.require.NoError(err)

	db, err := psqlutil.Connect(e.ctx, fmt.Sprintf( //nolint:nosprintfhostport
		"postgres://%s:%s@%s:%s/%s",
		dbCredential,
		dbCredential,
		host,
		port.Port(),
		dbCredential,
	))
	e.require.NoError(err)

	// run migrations
	sdb := stdlib.OpenDBFromPool(db.Pool)
	driver, err := pgx.WithInstance(sdb, &pgx.Config{}) //nolint:exhaustruct
	e.require.NoError(err)

	m, err := migrate.NewWithDatabaseInstance(
		"file://../migrations/",
		"pgxv5", driver,
	)
	e.require.NoError(err)

	e.require.NoError(m.Up())
	e.require.NoError(driver.Close())

	return db, stop
}

func (e *AppTestSuite) prepRedis() (*rdb.DB, stopFunc) {
	redisContainer, err := tcredis.Run(e.ctx, "redis:7.4-alpine")
	e.require.NoError(err)

	stop := func() { e.require.NoError(redisContainer.Terminate(e.ctx)) }

	uri, err := redisContainer.ConnectionString(e.ctx)
	e.require.NoError(err)

	connOpts, err := redis.ParseURL(uri)
	e.require.NoError(err)

	redis, err := rdb.Connect(e.ctx, connOpts.Addr, connOpts.Password, connOpts.DB)
	e.require.NoError(err)

	return redis, stop
}

func (e *AppTestSuite) getConfig() *config.Config {
	e.T().Setenv("APP_ENV", "test")
	e.T().Setenv("APP_URL", "localhost")
	e.T().Setenv("PASSWORD_SALT", "salty-password")
	e.T().Setenv("NOTE_PASSWORD_SALT", "salty-noted-password")
	e.T().Setenv("JWT_SIGNING_KEY", "jwt-key")
	e.T().Setenv("LOG_SHOW_LINE", "true")
	e.T().Setenv("LOG_FORMAT", "text")
	e.T().Setenv("LOG_LEVEL", "debug")

	return config.NewConfig()
}
