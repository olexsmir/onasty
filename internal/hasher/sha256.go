package hasher

import (
	"crypto/sha256"
	"encoding/hex"
)

type SHA256Hasher struct {
	salt string
}

func NewSHA256Hasher(salt string) *SHA256Hasher {
	return &SHA256Hasher{salt: salt}
}

func (h *SHA256Hasher) Hash(inp string) (string, error) {
	hash := sha256.New()
	if _, err := hash.Write([]byte(inp)); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum([]byte(h.salt))), nil
}
