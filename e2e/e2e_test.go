package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/olexsmir/onasty/internal/adapters/primary/web"
	"github.com/olexsmir/onasty/internal/adapters/secondary/store/psql"
	"github.com/olexsmir/onasty/internal/adapters/secondary/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/core/services/notesrv"
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

		postgresDB   *psql.DB
		stopPostgres stopDBFunc

		router http.Handler
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

func (s *AppTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.require = s.Require()

	db, stop, err := s.prepPostgres()
	s.Require().NoError(err)

	s.postgresDB = db
	s.stopPostgres = stop

	s.initDeps()
}

func (s *AppTestSuite) TearDownSuite() {
	s.stopPostgres()
}

func (s *AppTestSuite) prepPostgres() (*psql.DB, stopDBFunc, error) {
	// setup dcoker container
	dbCredential := "testing"
	postgresContainer, err := postgres.RunContainer(
		s.ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithUsername(dbCredential),
		postgres.WithPassword(dbCredential),
		postgres.WithDatabase(dbCredential),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp")),
	)
	s.require.NoError(err)

	stop := func() {
		err = postgresContainer.Terminate(s.ctx)
		s.require.NoError(err)
	}

	// connect to the db
	host, err := postgresContainer.Host(s.ctx)
	s.require.NoError(err)

	port, err := postgresContainer.MappedPort(s.ctx, "5432/tcp")
	s.require.NoError(err)

	db, err := psql.Connect(s.ctx, psql.Credentials{
		Username: dbCredential,
		Password: dbCredential,
		Host:     host,
		Port:     port.Port(),
		Database: dbCredential,
	})
	s.require.NoError(err)

	// run migrations
	sdb := stdlib.OpenDBFromPool(db.Pool)
	driver, err := pgx.WithInstance(sdb, &pgx.Config{})
	s.require.NoError(err)

	m, err := migrate.NewWithDatabaseInstance(
		"file://../migrations/",
		"pgxv5", driver,
	)
	s.require.NoError(err)

	err = m.Up()
	s.require.NoError(err)

	return db, stop, driver.Close()
}

func (s *AppTestSuite) initDeps() {
	noterepo := noterepo.New(s.postgresDB)
	notesrv := notesrv.New(noterepo)
	handlers := web.NewHandler(web.HandlerDeps{
		NoteService: notesrv,
	})

	s.router = handlers.InitRoutes()
}
