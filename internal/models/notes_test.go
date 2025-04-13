package models

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

func TestNote_Validate(t *testing.T) {
	// NOTE: there no need to test if note is expired since it tested in IsExpired test

	t.Run("should pass the validation only if content provided", func(t *testing.T) {
		n := Note{Content: "the content"} //nolint:exhaustruct
		assert.NoError(t, n.Validate())
	})
	t.Run("should pass validation with content and correct expiration time", func(t *testing.T) {
		n := Note{ //nolint:exhaustruct
			Content:   "content",
			ExpiresAt: time.Now().Add(time.Minute),
		}
		assert.NoError(t, n.Validate())
	})
	t.Run("should fail if content is missing", func(t *testing.T) {
		n := Note{Content: ""} //nolint:exhaustruct
		assert.EqualError(t, n.Validate(), ErrNoteContentIsEmpty.Error())
	})
	t.Run("should fail if content is missing and other fields are set", func(t *testing.T) {
		n := Note{ //nolint:exhaustruct
			Slug:                 "some-slug",
			Password:             "some-password",
			BurnBeforeExpiration: false,
		}
		assert.EqualError(t, n.Validate(), ErrNoteContentIsEmpty.Error())
	})
	t.Run("should fail if expiration time is in the past", func(t *testing.T) {
		n := Note{Content: "content", ExpiresAt: time.Now().Add(-time.Hour)} //nolint:exhaustruct
		assert.EqualError(t, n.Validate(), ErrNoteExpired.Error())
	})
}

func TestNote_IsExpired(t *testing.T) {
	t.Run("should be expired", func(t *testing.T) {
		note := Note{ExpiresAt: time.Now().Add(-time.Hour)} //nolint:exhaustruct
		assert.True(t, note.IsExpired())
	})
	t.Run("should be not expired", func(t *testing.T) {
		note := Note{ExpiresAt: time.Now().Add(time.Hour)} //nolint:exhaustruct
		assert.False(t, note.IsExpired())
	})
	t.Run("should be not expired when [ExpiredAt] is zero", func(t *testing.T) {
		note := Note{ExpiresAt: time.Time{}} //nolint:exhaustruct
		assert.False(t, note.IsExpired())
	})
}

func TestNote_ShouldBeBurnt(t *testing.T) {
	t.Run("should be burnt", func(t *testing.T) {
		note := Note{ //nolint:exhaustruct
			BurnBeforeExpiration: true,
			ExpiresAt:            time.Now().Add(time.Hour),
		}
		assert.True(t, note.ShouldBeBurnt())
	})
	t.Run("should not be burnt", func(t *testing.T) {
		note := Note{ //nolint:exhaustruct
			BurnBeforeExpiration: true,
			ExpiresAt:            time.Time{},
		}
		assert.False(t, note.ShouldBeBurnt())
	})
	t.Run("could not be burnt when expiration and shouldBurn set to false", func(t *testing.T) {
		note := Note{ //nolint:exhaustruct
			BurnBeforeExpiration: false,
			ExpiresAt:            time.Time{},
		}
		assert.False(t, note.ShouldBeBurnt())
	})
}

func TestNote_IsRead(t *testing.T) {
	t.Run("should be unread", func(t *testing.T) {
		n := Note{ReadAt: time.Time{}} //nolint:exhaustruct
		assert.False(t, n.IsRead())
	})
	t.Run("should be read", func(t *testing.T) {
		n := Note{ReadAt: time.Now()} //nolint:exhaustruct
		assert.True(t, n.IsRead())
	})
}
