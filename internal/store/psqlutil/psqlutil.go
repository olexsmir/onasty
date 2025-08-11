package psqlutil

import (
	"context"
	"database/sql"
	"errors"
	"time"

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
func IsDuplicateErr(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && // unique_violation
			pgErr.ConstraintName == constraintName
	}
	return false
}

// NullTimeToTime converts sql.NullTime to time.Time.
// Returns zero [time.Time] if NullTime is not valid.
func NullTimeToTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}
