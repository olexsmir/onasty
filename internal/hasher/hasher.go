package hasher

type Hasher interface {
	// Hash takes a string as input and returns its hash
	Hash(string) (string, error)
}
