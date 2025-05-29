package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

var usersData = []struct {
	id        string
	email     string
	password  string
	activated bool
}{
	{ //nolint:exhaustruct
		email:     "admin@onasty.local",
		password:  "adminadmin",
		activated: true,
	},
	{ //nolint:exhaustruct
		email:     "users@onasty.local",
		activated: false,
		password:  "qwerty123",
	},
}

func seedUsers(
	ctx context.Context,
	hash hasher.Hasher,
	db *psqlutil.DB,
) error {
	for i, user := range usersData {
		passwrd, err := hash.Hash(user.password)
		if err != nil {
			return err
		}

		var id pgtype.UUID
		err = db.QueryRow(ctx, `
			insert into users (email, password, activated, created_at, last_login_at)
			values ($1, $2, $3, $4, $5)
				on conflict (email) do update
				set password = excluded.password
			returning id::text
		`, user.email, passwrd, user.activated, time.Now(), time.Now()).
			Scan(&id)
		if err != nil {
			return err
		}

		usersData[i].id = id.String()
	}

	return nil
}
