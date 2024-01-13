package domain

import (
	"testing"
	"time"
)

func TestNote_Validate(t *testing.T) {
	tests := []struct {
		name    string
		fields  Note
		wantErr bool
	}{
		{
			name: "valid with slug",
			fields: Note{
				Content:   "content",
				Slug:      "slug",
				CreatedAt: time.Now(),
			},
		},
		{
			name: "valid without slug",
			fields: Note{
				Content:   "content",
				CreatedAt: time.Now(),
			},
		},
		{
			name: "valid with expires_at",
			fields: Note{
				Content:   "content",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
		},
		{
			name:    "invalid without content",
			wantErr: true,
			fields: Note{
				Content: "",
			},
		},
		{
			name:    "invalid with expires_at in past",
			wantErr: true,
			fields: Note{
				Content:   "content",
				ExpiresAt: time.Now().Add(-time.Hour),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Note{
				ID:        tt.fields.ID,
				Content:   tt.fields.Content,
				Slug:      tt.fields.Slug,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
			}
			if err := n.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Note.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
