package psql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	db, err := pgxpool.New(ctx, fmt.Sprintf( //nolint:nosprintfhostport
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
