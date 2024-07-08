package e2e

import (
	"github.com/henvic/pgq"
	"github.com/olexsmir/onasty/internal/models"
)

func (e *AppTestSuite) getUserFromDBByUsername(username string) models.User {
	query, args, err := pgq.
		Select("id", "username", "email", "password", "created_at", "last_login_at").
		From("users").
		Where(pgq.Eq{
			"username": username,
		}).
		SQL()
	e.require.NoError(err)

	var user models.User
	err = e.postgresDB.QueryRow(e.ctx, query, args...).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.LastLoginAt)
	e.require.NoError(err)

	return user
}
