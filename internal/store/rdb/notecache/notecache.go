package notecache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/store/rdb"
)

type NoteCacher interface {
	GetBySlug(ctx context.Context, slug string) (dtos.NoteMetadataDTO, error)
	SetBySlug(ctx context.Context, slug string, note dtos.NoteMetadataDTO) error
}

var _ NoteCacher = (*NoteCache)(nil)

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

func (n *NoteCache) GetBySlug(ctx context.Context, slug string) (dtos.NoteMetadataDTO, error) {
	cached, err := n.rdb.Get(ctx, "note:"+slug).Result()
	if err != nil {
		return dtos.NoteMetadataDTO{}, err
	}

	var mtd dtos.NoteMetadataDTO
	if err := json.Unmarshal([]byte(cached), &mtd); err != nil {
		return dtos.NoteMetadataDTO{}, err
	}

	return mtd, nil
}

func (n *NoteCache) SetBySlug(ctx context.Context, slug string, note dtos.NoteMetadataDTO) error {
	nj, err := json.Marshal(note)
	if err != nil {
		return err
	}

	_, err = n.rdb.
		Set(ctx, "note:"+slug, nj, n.ttl).
		Result()

	return err
}
