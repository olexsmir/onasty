package models

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

//nolint:exhaustruct
func TestNote_Validate(t *testing.T) {
	// NOTE: there no need to test if note is expired since it tested in IsExpired test

	t.Run("should pass the validation if only slug and content are provided", func(t *testing.T) {
		n := Note{Content: "the content", Slug: "s"}
		assert.NoError(t, n.Validate())
	})
	t.Run("should pass validation with content and correct expiration time", func(t *testing.T) {
		n := Note{
			Content:   "content",
			Slug:      "s",
			ExpiresAt: time.Now().Add(time.Minute),
		}
		assert.NoError(t, n.Validate())
	})
	t.Run("should fail if content is missing", func(t *testing.T) {
		n := Note{Content: ""}
		assert.EqualError(t, n.Validate(), ErrNoteContentIsEmpty.Error())
	})
	t.Run("should fail if content is missing and other fields are set", func(t *testing.T) {
		n := Note{
			Slug:                 "some-slug",
			Password:             "some-password",
			BurnBeforeExpiration: false,
		}
		assert.EqualError(t, n.Validate(), ErrNoteContentIsEmpty.Error())
	})
	t.Run("should fail if expiration time is in the past", func(t *testing.T) {
		n := Note{
			Content:   "content",
			Slug:      "s",
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		assert.EqualError(t, n.Validate(), ErrNoteExpired.Error())
	})
	t.Run("should fail if burn before expiration is set, and expiration time is not",
		func(t *testing.T) {
			n := Note{
				Content:              "content",
				BurnBeforeExpiration: true,
			}

			assert.EqualError(t, n.Validate(), ErrNoteCannotBeBurnt.Error())
		},
	)
	t.Run("should fail if slug is empty", func(t *testing.T) {
		n := Note{Content: "the content", Slug: " "}
		assert.EqualError(t, n.Validate(), ErrNoteSlugIsInvalid.Error())
	})
	t.Run("should fail if slug has '/'", func(t *testing.T) {
		n := Note{Content: "the content", Slug: "asdf/asdf"}
		assert.EqualError(t, n.Validate(), ErrNoteSlugIsInvalid.Error())
	})
	t.Run("should fail if slug one of not allowed slugs", func(t *testing.T) {
		for notAllowedSlug := range notAllowedSlugs {
			n := Note{Content: "the content", Slug: notAllowedSlug}
			assert.EqualError(t, n.Validate(), ErrNoteSlugIsAlreadyInUse.Error())
		}
	})
}

//nolint:exhaustruct
func TestNote_IsExpired(t *testing.T) {
	t.Run("should be expired", func(t *testing.T) {
		note := Note{ExpiresAt: time.Now().Add(-time.Hour)}
		assert.True(t, note.IsExpired())
	})
	t.Run("should be not expired", func(t *testing.T) {
		note := Note{ExpiresAt: time.Now().Add(time.Hour)}
		assert.False(t, note.IsExpired())
	})
	t.Run("should be not expired when [ExpiredAt] is zero", func(t *testing.T) {
		note := Note{ExpiresAt: time.Time{}}
		assert.False(t, note.IsExpired())
	})
}

//nolint:exhaustruct
func TestNote_ShouldBeBurnt(t *testing.T) {
	t.Run("should be burnt", func(t *testing.T) {
		note := Note{
			BurnBeforeExpiration: true,
			ExpiresAt:            time.Now().Add(time.Hour),
		}
		assert.True(t, note.ShouldBeBurnt())
	})
	t.Run("should not be burnt", func(t *testing.T) {
		note := Note{
			BurnBeforeExpiration: true,
			ExpiresAt:            time.Time{},
		}
		assert.False(t, note.ShouldBeBurnt())
	})
	t.Run("could not be burnt when expiration and shouldBurn set to false", func(t *testing.T) {
		note := Note{
			BurnBeforeExpiration: false,
			ExpiresAt:            time.Time{},
		}
		assert.False(t, note.ShouldBeBurnt())
	})
}

//nolint:exhaustruct
func TestNote_IsRead(t *testing.T) {
	t.Run("should be unread", func(t *testing.T) {
		n := Note{ReadAt: time.Time{}}
		assert.False(t, n.IsRead())
	})
	t.Run("should be read", func(t *testing.T) {
		n := Note{ReadAt: time.Now()}
		assert.True(t, n.IsRead())
	})
}
