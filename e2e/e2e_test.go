package e2e_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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
	"github.com/olexsmir/onasty/internal/mailer"
	"github.com/olexsmir/onasty/internal/service/notesrv"
	"github.com/olexsmir/onasty/internal/service/usersrv"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/store/psql/sessionrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psql/vertokrepo"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
	"github.com/olexsmir/onasty/internal/store/rdb/usercache"
	httptransport "github.com/olexsmir/onasty/internal/transport/http"
	"github.com/olexsmir/onasty/internal/transport/http/ratelimit"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tsredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type (
	stopDBFunc   func()
	AppTestSuite struct {
		suite.Suite

		ctx     context.Context
		require *require.Assertions

		postgresDB   *psqlutil.DB
		stopPostgres stopDBFunc

		redis     *redis.Client
		stopRedis stopDBFunc

		router       http.Handler
		hasher       hasher.Hasher
		jwtTokenizer jwtutil.JWTTokenizer
		mailer       *mailer.TestMailer
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

	db, stop := e.prepPostgres()
	e.postgresDB = db
	e.stopPostgres = stop

	rdb, stop := e.prepRedis()
	e.redis = rdb
	e.stopRedis = stop

	e.initDeps()
}

func (e *AppTestSuite) TearDownSuite() {
	e.stopPostgres()
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
	e.mailer = mailer.NewTestMailer()

	sessionrepo := sessionrepo.New(e.postgresDB)
	vertokrepo := vertokrepo.New(e.postgresDB)

	userepo := userepo.New(e.postgresDB)
	usercache := usercache.New(e.redis, cfg.CacheUsersTTL)
	usersrv := usersrv.New(
		userepo,
		sessionrepo,
		vertokrepo,
		e.hasher,
		e.jwtTokenizer,
		e.mailer,
		usercache,
		cfg.JwtRefreshTokenTTL,
		cfg.VerificationTokenTTL,
		cfg.AppURL,
	)

	noterepo := noterepo.New(e.postgresDB)
	notesrv := notesrv.New(noterepo)

	// for testing purposes, it's ok to have high values ig
	ratelimitCfg := ratelimit.Config{
		RPS:   1000,
		TTL:   time.Millisecond,
		Burst: 1000,
	}

	handler := httptransport.NewTransport(usersrv, notesrv, ratelimitCfg)
	e.router = handler.Handler()
}

func (e *AppTestSuite) prepPostgres() (*psqlutil.DB, stopDBFunc) {
	dbCredential := "testing"
	postgresContainer, err := postgres.Run(e.ctx,
		"postgres:16-alpine",
		postgres.WithUsername(dbCredential),
		postgres.WithPassword(dbCredential),
		postgres.WithDatabase(dbCredential),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")))
	e.require.NoError(err)

	stop := func() { e.require.NoError(postgresContainer.Terminate(e.ctx)) }

	// connect to the db
	host, err := postgresContainer.Host(e.ctx)
	e.require.NoError(err)

	port, err := postgresContainer.MappedPort(e.ctx, "5432/tcp")
	e.require.NoError(err)

	db, err := psqlutil.Connect(
		e.ctx,
		fmt.Sprintf( //nolint:nosprintfhostport
			"postgres://%s:%s@%s:%s/%s",
			dbCredential,
			dbCredential,
			host,
			port.Port(),
			dbCredential,
		),
	)
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

func (e *AppTestSuite) prepRedis() (*redis.Client, stopDBFunc) {
	redisContainer, err := tsredis.Run(e.ctx, "redis:7.4-alpine")
	e.require.NoError(err)

	stop := func() { e.require.NoError(redisContainer.Terminate(e.ctx)) }

	uri, err := redisContainer.ConnectionString(e.ctx)
	e.require.NoError(err)

	connOpts, err := redis.ParseURL(uri)
	e.require.NoError(err)

	rdb := redis.NewClient(connOpts)
	return rdb, stop
}

func (e *AppTestSuite) getConfig() *config.Config {
	return &config.Config{ //nolint:exhaustruct
		AppEnv:               "testing",
		AppURL:               "",
		ServerPort:           "3000",
		PasswordSalt:         "salty-password",
		JwtSigningKey:        "jwt-key",
		JwtAccessTokenTTL:    time.Hour,
		JwtRefreshTokenTTL:   24 * time.Hour,
		VerificationTokenTTL: 24 * time.Hour,
		LogShowLine:          os.Getenv("LOG_SHOW_LINE") == "true",
		LogFormat:            "text",
		LogLevel:             "debug",
	}
}
