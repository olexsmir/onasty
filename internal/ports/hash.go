package ports

type Hasher interface {
	// Hash returns a hashed string
	Hash(string) ([]byte, error)

	// Compare compares two hashed string for matching
	Compare([]byte, []byte) (bool, error)
}
