package notecache

import (
	"bytes"
	"context"
	"encoding/gob"
	"strings"
	"time"

	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/rdb"
)

type NoteCacher interface {
	SetNote(ctx context.Context, slug string, note models.Note) error
	GetNote(ctx context.Context, slug string) (models.Note, error)
}

type NoteCache struct {
	rdb *rdb.DB
	ttl time.Duration
}

func New(rdb *rdb.DB, ttl time.Duration) *NoteCache {
	return &NoteCache{
		rdb: rdb,
		ttl: ttl,
	}
}

func (n *NoteCache) SetNote(ctx context.Context, slug string, note models.Note) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(note); err != nil {
		return err
	}

	_, err := n.rdb.Set(ctx, getKey(slug), buf.Bytes(), n.ttl).Result()
	return err
}

func (n *NoteCache) GetNote(ctx context.Context, slug string) (models.Note, error) {
	val, err := n.rdb.Get(ctx, getKey(slug)).Bytes()
	if err != nil {
		return models.Note{}, err
	}

	var note models.Note
	if err = gob.NewDecoder(bytes.NewReader(val)).Decode(&note); err != nil {
		return models.Note{}, err
	}

	return note, err
}

func getKey(slug string) string {
	var sb strings.Builder
	sb.WriteString("note:")
	sb.WriteString(slug)
	return sb.String()
}
