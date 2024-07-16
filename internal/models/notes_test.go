package models

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

func TestNote_Validate(t *testing.T) {
	tests := []struct {
		name      string
		note      Note
		willError bool
		error     error
	}{
		// NOTE: there no need to test if note is expired since it tested in IsExpired test
		{
			name: "ok",
			note: Note{
				Content:   "some wired ass content",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			willError: false,
		},
		{
			name:      "content missing",
			note:      Note{Content: ""},
			willError: true,
			error:     ErrNoteContentIsEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.note.Validate()
			if tt.willError {
				assert.EqualError(t, err, tt.error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNote_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		note     Note
		expected bool
	}{
		{
			name:     "expired",
			note:     Note{ExpiresAt: time.Now().Add(-time.Hour)},
			expected: true,
		},
		{
			name:     "not expired",
			note:     Note{ExpiresAt: time.Now().Add(time.Hour)},
			expected: false,
		},
		{
			name:     "zero expiration",
			note:     Note{ExpiresAt: time.Time{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.note.IsExpired())
		})
	}
}
