package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
