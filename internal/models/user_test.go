package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name string
		fail bool

		username string
		email    string
		password string
	}{
		{
			name:     "valid",
			fail:     false,
			email:    "test@example.org",
			username: "iuserarchbtw",
			password: "superhardasspassword",
		},
		{
			name:     "all fields empty",
			fail:     true,
			email:    "",
			username: "",
			password: "",
		},
		{
			name:     "invalid email",
			fail:     true,
			email:    "test",
			username: "iuserarchbtw",
			password: "superhardasspassword",
		},
		{
			name:     "invalid password",
			fail:     true,
			email:    "test@example.org",
			username: "iuserarchbtw",
			password: "12345",
		},
		{
			name:     "invalid username",
			fail:     true,
			email:    "test@example.org",
			username: "",
			password: "superhardasspassword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := User{ //nolint:exhaustruct
				Username: tt.username,
				Email:    tt.email,
				Password: tt.password,
			}.Validate()

			if tt.fail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
