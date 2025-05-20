package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name string
		fail bool

		email    string
		password string
	}{
		{
			name:     "valid",
			fail:     false,
			email:    "test@example.org",
			password: "superhardasspassword",
		},
		{
			name:     "all fields empty",
			fail:     true,
			email:    "",
			password: "",
		},
		{
			name:     "invalid email",
			fail:     true,
			email:    "test",
			password: "superhardasspassword",
		},
		{
			name:     "invalid password",
			fail:     true,
			email:    "test@example.org",
			password: "12345",
		},
		{
			name:     "invalid username",
			fail:     true,
			email:    "test@example.org",
			password: "superhardasspassword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := User{ //nolint:exhaustruct
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
