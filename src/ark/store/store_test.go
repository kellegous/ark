package store

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type testStore struct {
	Store
	dir string
}

func (s *testStore) Close() error {
	err := s.Store.Close()
	if err := os.RemoveAll(s.dir); err != nil {
		return err
	}
	return err
}

func openTestStore(t *testing.T) Store {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	s, err := Open(filepath.Join(tmp, "r.db"))
	if err != nil {
		t.Fatal(err)
	}

	return &testStore{
		Store: s,
		dir:   tmp,
	}
}

func TestOpenClose(t *testing.T) {
	c := openTestStore(t)
	defer c.Close()
}

func sameStringArrays(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, n := 0, len(a); i < n; i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func sameRoute(a, b *Route) bool {
	return a.Name == b.Name &&
		a.Port == b.Port &&
		sameStringArrays(a.Hosts, b.Hosts) &&
		sameStringArrays(a.Backends, b.Backends)
}

func TestSaveLoad(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()

	a := Route{
		Name:  "foo",
		Port:  222,
		Hosts: []string{"a", "b"},
	}

	if err := s.Save(&a); err != nil {
		t.Fatal(err)
	}

	var b Route
	if err := s.Load("foo", &b); err != nil {
		t.Fatal(err)
	}

	if !sameRoute(&a, &b) {
		t.Fatalf("expected %v got %v", &a, &b)
	}
}

func TestLoadAll(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()

	routes := map[string]*Route{
		"foo": &Route{
			Name:  "foo",
			Port:  2222,
			Hosts: []string{"a"},
		},

		"bar": &Route{
			Name:  "bar",
			Port:  2228,
			Hosts: []string{"z"},
		},

		"baz": &Route{
			Name:  "baz",
			Port:  80,
			Hosts: []string{"y"},
		},
	}

	for _, route := range routes {
		if err := s.Save(route); err != nil {
			t.Fatal(err)
		}
	}

	rts, err := s.LoadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(rts) != len(routes) {
		t.Fatalf("not enough results: expected %d got %d", len(routes), len(rts))
	}

	for _, rt := range rts {
		if !sameRoute(rt, routes[rt.Name]) {
			t.Fatalf("expected %v got %v", rt, routes[rt.Name])
		}
	}
}

func TestDelete(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()

	if err := s.Save(&Route{
		Name:  "foo",
		Port:  80,
		Hosts: []string{"foo"},
	}); err != nil {
		t.Fatal(err)
	}

	rts, err := s.LoadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(rts) != 1 {
		t.Fatalf("expected 1 route got %d", len(rts))
	}

	if err := s.Delete("foo"); err != nil {
		t.Fatal(err)
	}

	rts, err = s.LoadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(rts) != 0 {
		t.Fatalf("expected no routes got %d", len(rts))
	}
}
