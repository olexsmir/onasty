package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

var notesData = []struct {
	id                   string
	content              string
	slug                 string
	keepBeforeExpiration bool
	password             string
	expiresAt            time.Time
	hasAuthor            bool
	authorID             int
}{
	{ //nolint:exhaustruct
		content:              "that test note one",
		slug:                 "one",
		keepBeforeExpiration: false,
	},
	{ //nolint:exhaustruct
		content:              "that test note two",
		slug:                 "two",
		keepBeforeExpiration: true,
		password:             "",
		expiresAt:            time.Now().Add(24 * time.Hour),
	},
	{ //nolint:exhaustruct
		content:  "that passworded note",
		slug:     "passwd",
		password: "pass",
	},
	{ //nolint:exhaustruct
		content:   "that note with author",
		slug:      "user",
		hasAuthor: true,
		authorID:  0,
	},
	{ //nolint:exhaustruct
		content:   "that another authored note",
		slug:      "user2",
		hasAuthor: true,
		authorID:  0,
	},
	{ //nolint:exhaustruct
		content:   "that another authored note",
		slug:      "user3",
		password:  "passwd",
		hasAuthor: true,
		authorID:  0,
	},
}

func seedNotes(
	ctx context.Context,
	hash hasher.Hasher,
	db *psqlutil.DB,
) error {
	for i, note := range notesData {
		passwd := ""
		if note.password != "" {
			var err error
			passwd, err = hash.Hash(note.password)
			if err != nil {
				return err
			}
		}

		err := db.QueryRow(
			ctx, `
		insert into notes (content, slug, keep_before_expiration, password, expires_at)
		values ($1, $2, $3, $4, $5)
		on conflict (slug) do update set
			content = excluded.content,
			keep_before_expiration = excluded.keep_before_expiration,
			password = excluded.password,
			expires_at = excluded.expires_at
		returning id`,
			note.content,
			note.slug,
			note.keepBeforeExpiration,
			passwd,
			note.expiresAt,
		).Scan(&notesData[i].id)
		if err != nil {
			return err
		}

		if note.hasAuthor {
			slog.Info("setting author", "note", note.id, "author", note.authorID)
			if err := setAuthor(ctx, db, notesData[i].id, usersData[note.authorID].id); err != nil {
				return err
			}
		}
	}

	return nil
}

func setAuthor(
	ctx context.Context,
	db *psqlutil.DB,
	noteID string,
	authorID string,
) error {
	_, err := db.Exec(
		ctx,
		`insert into notes_authors (note_id, user_id) values ($1, $2)`,
		noteID, authorID)
	return err
}
