package store

import (
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
)

// ErrNotFound ...
var ErrNotFound = leveldb.ErrNotFound

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

// Save ...
func (s *Store) Save(r *Route) error {
	b, err := proto.Marshal(r)
	if err != nil {
		return err
	}

	return s.db.Put([]byte(r.Name), b, nil)
}

// Load ...
func (s *Store) Load(name string, r *Route) error {
	b, err := s.db.Get([]byte(name), nil)
	if err != nil {
		return err
	}

	return proto.Unmarshal(b, r)
}

// LoadAll ...
func (s *Store) LoadAll() ([]*Route, error) {
	var rts []*Route

	it := s.db.NewIterator(nil, nil)
	defer it.Release()

	for it.Next() {
		r := &Route{}

		if err := proto.Unmarshal(it.Value(), r); err != nil {
			return nil, err
		}

		rts = append(rts, r)
	}

	if err := it.Error(); err != nil {
		return nil, err
	}

	return rts, nil
}

// Close ...
func (s *Store) Close() error {
	return s.db.Close()
}
