package port

// HashPort is the outbound port for password hashing operations.
//
//go:generate mockgen -source=hash_port.go -destination=mocks/mock_hash_port.go -package=mocks
type HashPort interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}
