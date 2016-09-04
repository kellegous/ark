package api

import (
	"testing"

	"ark/store"
)

type mockLoadBalancer struct {
	count int
}

type mockStore struct {
	rts map[string]*store.Route
}

func (l *mockLoadBalancer) Update([]*store.Route) error {
	l.count++
	return nil
}

func newStore() store.Store {
	return &mockStore{
		rts: map[string]*store.Route{},
	}
}

func (s *mockStore) Save(r *store.Route) error {
	s.rts[r.Name] = r
	return nil
}

func (s *mockStore) Load(name string, r *store.Route) error {
	rt := s.rts[name]
	if rt == nil {
		return store.ErrNotFound
	}
	*r = *rt
	return nil
}

func (s *mockStore) LoadAll() ([]*store.Route, error) {
	var rts []*store.Route
	for _, rt := range s.rts {
		rts = append(rts, rt)
	}
	return rts, nil
}

func (s *mockStore) Delete(name string) error {
	if s.rts[name] == nil {
		return store.ErrNotFound
	}
	delete(s.rts, name)
	return nil
}

func (s *mockStore) Close() error {
	return nil
}

func TestPostRoutes(t *testing.T) {
	// TODO(knorton): Test this.
}
