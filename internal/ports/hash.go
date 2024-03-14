package ports

type Hasher interface {
	// Hash returns a hashed string
	Hash(string) (string, error)
}
