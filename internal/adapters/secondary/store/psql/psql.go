package psql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	pgxuuid "github.com/vgarvardt/pgx-google-uuid/v5"
)

type Credentials struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
}

type DB struct{ *pgxpool.Pool }

func Connect(ctx context.Context, conn Credentials) (*DB, error) {
	dbcfg, err := pgxpool.ParseConfig(fmt.Sprintf( //nolint:nosprintfhostport
		"postgres://%s:%s@%s:%s/%s",
		conn.Username,
		conn.Password,
		conn.Host,
		conn.Port,
		conn.Database,
	))
	if err != nil {
		return nil, err
	}

	dbcfg.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		return nil
	}

	db, err := pgxpool.NewWithConfig(ctx, dbcfg)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	return &DB{
		Pool: db,
	}, nil
}

func (d *DB) Close() error {
	d.Pool.Close()
	return nil
}

func IsDuplicateErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}
