package hasher

type Hasher interface {
	Hash(data string) (string, error)
	Compare(data, encodedHash string) (bool, error)
}
