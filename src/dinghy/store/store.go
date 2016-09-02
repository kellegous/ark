package store

import (
	"github.com/syndtr/goleveldb/leveldb"
)

// Store ...
type Store struct {
	db *leveldb.DB
}

// Open ...
func Open(path string) (*Store, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	return &Store{
		db: db,
	}, nil
}

// Close ...
func (s *Store) Close() error {
	return s.db.Close()
}
