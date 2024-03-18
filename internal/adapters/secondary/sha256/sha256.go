package sha256

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/olexsmir/onasty/internal/ports"
)

var _ ports.Hasher = (*SHA256Hasher)(nil)

type SHA256Hasher struct { //nolint:revive
	salt string
}

func NewSHA256Hasher(salt string) ports.Hasher {
	return &SHA256Hasher{salt: salt}
}

func (h *SHA256Hasher) Hash(inp string) (string, error) {
	hash := sha256.New()
	if _, err := hash.Write([]byte(inp)); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum([]byte(h.salt))), nil
}
