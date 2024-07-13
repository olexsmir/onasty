package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/jwtutil"
	"github.com/olexsmir/onasty/internal/service/usersrv"
	"github.com/olexsmir/onasty/internal/store/psql/sessionrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
	httptransport "github.com/olexsmir/onasty/internal/transport/http"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
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

		router       http.Handler
		hasher       hasher.Hasher
		jwtTokenizer jwtutil.JWTTokenizer
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

	db, stop, err := e.prepPostgres()
	e.Require().NoError(err)

	e.postgresDB = db
	e.stopPostgres = stop

	e.initDeps()
}

func (e *AppTestSuite) TearDownSuite() {
	e.stopPostgres()
}

// initDeps initializes the dependencies for the app
// and sets up the router for tests
func (e *AppTestSuite) initDeps() {
	e.hasher = hasher.NewSHA256Hasher("pass_salt")
	e.jwtTokenizer = jwtutil.NewJWTUtil("jwt", time.Hour)

	sessionrepo := sessionrepo.New(e.postgresDB)

	userepo := userepo.New(e.postgresDB)
	usersrv := usersrv.New(userepo, sessionrepo, e.hasher, e.jwtTokenizer)

	handler := httptransport.NewTransport(usersrv)
	e.router = handler.Handler()
}

func (e *AppTestSuite) prepPostgres() (*psqlutil.DB, stopDBFunc, error) {
	dbCredential := "testing"
	postgresContainer, err := postgres.RunContainer(
		e.ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithUsername(dbCredential),
		postgres.WithPassword(dbCredential),
		postgres.WithDatabase(dbCredential),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp")),
	)
	e.require.NoError(err)

	stop := func() {
		err = postgresContainer.Terminate(e.ctx)
		e.require.NoError(err)
	}

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
	driver, err := pgx.WithInstance(sdb, &pgx.Config{})
	e.require.NoError(err)

	m, err := migrate.NewWithDatabaseInstance(
		"file://../migrations/",
		"pgxv5", driver,
	)
	e.require.NoError(err)

	err = m.Up()
	e.require.NoError(err)

	return db, stop, driver.Close()
}
