package router

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"testing"
)

type handler struct {
	tags  map[string]bool
	names [][]string
}

func (h *handler) handler(tag string) Handler {
	return func(w http.ResponseWriter, r *http.Request, names []string) {
		h.tags[tag] = true
		h.names = append(h.names, names)
	}
}

func (h *handler) taggedWith(tags ...string) bool {
	if len(h.tags) != len(tags) {
		return false
	}

	for _, tag := range tags {
		if !h.tags[tag] {
			return false
		}
	}

	return true
}

func (h *handler) lastNames() []string {
	if len(h.names) == 0 {
		return nil
	}
	return h.names[len(h.names)-1]
}

func (h *handler) clear() {
	h.tags = map[string]bool{}
	h.names = nil
}

type respWriter struct {
	bytes.Buffer
	status int
	header http.Header
}

func (w *respWriter) Header() http.Header {
	return w.header
}

func (w *respWriter) WriteHeader(status int) {
	w.status = status
}

func dispatch(h http.Handler, verb, path string) int {
	req, err := http.NewRequest(verb, fmt.Sprintf("http:%s", path), nil)
	if err != nil {
		log.Panic(err)
	}

	res := respWriter{
    header: http.Header(map[string][]string{}),
  }

	h.ServeHTTP(&res, req)

	if s := res.status; s != 0 {
		return s
	}

	return http.StatusOK
}

func expectDispatch(
	t *testing.T,
	status int,
	h http.Handler,
	verb, path string) {
	if s := dispatch(h, verb, path); s != status {
		t.Fatalf("Expected status %d, got %d (%s, %s)", status, s, verb, path)
	}
}

func TestPaths(t *testing.T) {
	r := New()

	rec := handler{tags: map[string]bool{}}

	r.Handle(Get, "/a", rec.handler("/a"))
	r.Handle(Get, "/a/*", rec.handler("/a/*"))
	r.Handle(Get, "/a/", rec.handler("/a/"))
	r.Handle(Get, "/a/b/c/*", rec.handler("/a/b/c/*"))

	h := r.Build()

	expectDispatch(t, http.StatusOK, h, "GET", "/a")
	if !rec.taggedWith("/a") {
		t.Fatalf("expected only /a to be called: %v", rec)
	}
	rec.clear()

	expectDispatch(t, http.StatusOK, h, "GET", "/a/b")
	if !rec.taggedWith("/a/*") {
		t.Fatalf("expected only /a/* to be called: %v", rec)
	}
	rec.clear()

	expectDispatch(t, http.StatusOK, h, "GET", "/a/")
	if !rec.taggedWith("/a/") {
		t.Fatalf("expected only /a/ to be called: %v", rec)
	}
	rec.clear()

	expectDispatch(t, http.StatusOK, h, "GET", "/a/b/c/ddef")
	if !rec.taggedWith("/a/b/c/*") {
		t.Fatalf("expected only /a/b/c/* to be called: %v", rec)
	}
	rec.clear()

	expectDispatch(t, http.StatusNotFound, h, "GET", "/")
	if !rec.taggedWith() {
		t.Fatalf("expected no call: %v", rec)
	}
	rec.clear()
}
