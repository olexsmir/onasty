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
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type stopDBFunc func()

type AppTestSuite struct {
	suite.Suite

	ctx    context.Context
	db     *psql.DB
	stopDB stopDBFunc

	router http.Handler
}

func TestAppSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// gin output is too verbose(and annoying) in tests
	gin.SetMode(gin.TestMode)

	suite.Run(t, new(AppTestSuite))
}

func (s *AppTestSuite) SetupTest() {
	s.ctx = context.Background()

	db, stop, err := s.prepPostgres()
	s.Require().NoError(err)

	s.db = db
	s.stopDB = stop

	s.initDeps()
}

func (s *AppTestSuite) TearDownSuite() {
	s.stopDB()
}

func (s *AppTestSuite) prepPostgres() (*psql.DB, stopDBFunc, error) {
	postgresContainer, err := postgres.RunContainer(
		s.ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithUsername("testing"),
		postgres.WithPassword("testing"),
		postgres.WithDatabase("testing"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp")))
	if err != nil {
		return nil, func() {}, err
	}

	stop := func() { _ = postgresContainer.Terminate(s.ctx) }

	host, err := postgresContainer.Host(s.ctx)
	if err != nil {
		return nil, func() {}, err
	}

	port, err := postgresContainer.MappedPort(s.ctx, "5432/tcp")
	if err != nil {
		return nil, func() {}, err
	}

	db, err := psql.Connect(s.ctx, psql.Credentials{
		Username: "testing",
		Password: "testing",
		Host:     host,
		Port:     port.Port(),
		Database: "testing",
	})
	if err != nil {
		return nil, func() {}, err
	}

	sdb := stdlib.OpenDBFromPool(db.Pool)
	driver, err := pgx.WithInstance(sdb, &pgx.Config{})
	if err != nil {
		return nil, func() {}, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../migrations/",
		"pgxv5", driver,
	)
	if err != nil {
		return nil, func() {}, err
	}

	if err := m.Up(); err != nil {
		return nil, func() {}, err
	}

	return db, stop, driver.Close()
}

func (s *AppTestSuite) initDeps() {
	noterepo := noterepo.New(s.db)
	notesrv := notesrv.New(noterepo)
	handlers := web.NewHandler(web.HandlerDeps{
		NoteService: notesrv,
	})

	s.router = handlers.InitRoutes()
}
