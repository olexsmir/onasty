package psqlutil

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct{ *pgxpool.Pool }

func Connect(ctx context.Context, dsn string) (*DB, error) {
	db, err := pgxpool.New(ctx, dsn)
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
