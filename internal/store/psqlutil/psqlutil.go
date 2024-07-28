package psqlutil

import (
	"context"
	"errors"

	"github.com/jackc/pgconn"
	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct{ *pgxpool.Pool }

func Connect(ctx context.Context, dsn string) (*DB, error) {
	dbConf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	dbConf.AfterConnect = func(_ context.Context, c *pgx.Conn) error {
		pgxuuid.Register(c.TypeMap())
		return nil
	}

	db, err := pgxpool.NewWithConfig(ctx, dbConf)
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

func (db *DB) Close() error {
	db.Pool.Close()
	return nil
}

// IsDuplicateErr function that checks if the error is a duplicate key violation.
func IsDuplicateErr(err error, constraintName ...string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if len(constraintName) == 0 || len(constraintName) == 1 {
			return pgErr.Code == "23505" && // unique_violation
				pgErr.ConstraintName == constraintName[0]
		}
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}
