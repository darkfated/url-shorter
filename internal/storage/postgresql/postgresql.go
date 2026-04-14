package postgresql

type Store struct{}

func New(dsn string) (*Store, error) {
	_ = dsn
	return &Store{}, nil
}
