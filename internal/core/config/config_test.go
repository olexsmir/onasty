package config

import "testing"

func TestGetenvOrDefault(t *testing.T) {
	t.Run("should return default value if env variable is not set", func(t *testing.T) {
		def := "3000"
		serverPort := GetenvOrDefault("SERVER_PORT", def)

		if serverPort != def {
			t.Errorf("GetenvOrDefault() = %v, want %v", serverPort, def)
		}
	})

	t.Run("should return env variable if set", func(t *testing.T) {
		userPort := "4000"
		t.Setenv("SERVER_PORT", userPort)

		if p := GetenvOrDefault("SERVER_PORT", "3000"); p != userPort {
			t.Errorf("GetenvOrDefault() = %v, want %v", p, userPort)
		}
	})
}
