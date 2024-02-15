package argon2

import (
	"github.com/alexedwards/argon2id"
	"github.com/olexsmir/onasty/internal/ports"
)

var _ ports.Hasher = (*Hasher)(nil)

type Hasher struct{}

func New() ports.Hasher {
	return &Hasher{}
}

func (h *Hasher) Hash(inp string) ([]byte, error) {
	hash, err := argon2id.CreateHash(inp, argon2id.DefaultParams)
	return []byte(hash), err
}

func (h *Hasher) Compare(firstInp []byte, secondInp []byte) (bool, error) {
	return argon2id.ComparePasswordAndHash(string(firstInp), string(secondInp))
}
