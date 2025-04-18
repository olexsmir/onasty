package hasher

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSHA256Hasher_Hash(t *testing.T) {
	hasher := NewSHA256Hasher("salt")

	hashed, err := hasher.Hash("qwerty123")
	require.NoError(t, err)
	require.NotEmpty(t, hashed)
}

func TestSHA256Hasher_Compared(t *testing.T) {
	hasher := NewSHA256Hasher("salt")
	input := "qwerty123"

	t.Run("valid", func(t *testing.T) {
		hashed, err := hasher.Hash(input)
		require.NoError(t, err)
		require.NotEmpty(t, hashed)

		err = hasher.Compare(hashed, input)
		require.NoError(t, err)
	})

	t.Run("hashes mismatch", func(t *testing.T) {
		hashed, err := hasher.Hash(input + "4")
		require.NoError(t, err)
		require.NotEmpty(t, hashed)

		err = hasher.Compare(hashed, input)
		require.ErrorIs(t, err, ErrMismatchedHashes)
	})
}
