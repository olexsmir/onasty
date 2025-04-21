package hasher

import "errors"

var ErrMismatchedHashes = errors.New("hashes are mismatched")

type Hasher interface {
	// Hash takes a string as input and returns its hash
	Hash(str string) (string, error)

	// Compare takes two hashes and compares them
	// in case of mismatch returns [ErrMismatchedHashes]
	Compare(hash, plain string) error
}
