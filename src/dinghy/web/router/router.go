package router

import (
	"net/http"
	"strings"
)

// Verb ...
type Verb int

const (
	// Delete ...
	Delete Verb = iota
	// Get ...
	Get
	// Head ...
	Head
	// Options ...
	Options
	// Patch ...
	Patch
	// Post ...
	Post
	// Put ...
	Put
	unknownVerb
)

var verbs = map[string]Verb{
	"DELETE":  Delete,
	"GET":     Get,
	"HEAD":    Head,
	"OPTIONS": Options,
	"PATCH":   Patch,
	"POST":    Post,
	"PUT":     Put,
}

// Handler ...
type Handler func(http.ResponseWriter, *http.Request, []string)

// Builder ...
type Builder interface {
	Handle(Verb, string, Handler)
  HandleAll(string, Handler)
	Build() http.Handler
}

type router struct {
	ch map[string]*router
	rt *route
}

type route struct {
	vb [unknownVerb]Handler
}

func (r *router) place(path string) *router {
	if path == "" {
		return r
	}

	if r.ch == nil {
		r.ch = map[string]*router{}
	}

	k := path
	ix := strings.Index(path, "/")
	if ix >= 0 {
		k = path[:ix+1]
	}

	ch := r.ch[k]
	if ch == nil {
		ch = &router{}
		r.ch[k] = ch
	}

	return ch.place(path[len(k):])
}

func (r *router) find(path string, names *[]string) *router {
	if path == "" {
		return r
	}

	k := path
	ix := strings.Index(path, "/")
	if ix >= 0 {
		k = path[:ix+1]
	}

	if c := r.ch[k]; c != nil {
		if h := c.find(path[len(k):], names); h != nil {
			if h.rt != nil {
				return h
			}
		}
	}

	w := "*"
	if k[len(k)-1] == '/' {
		w = "*/"
	}

	if c := r.ch[w]; c != nil {
		*names = append(*names, strings.TrimRight(k, "/"))
		if h := c.find(path[len(k):], names); h != nil {
			if h.rt != nil {
				return h
			}
		}
	}

	return nil
}

func (r *router) set(verb Verb, h Handler) {
	if r.rt == nil {
		r.rt = &route{}
	}
	r.rt.vb[verb] = h
}

// ServeHTTP ...
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	v, ok := verbs[req.Method]
	if !ok {
		http.Error(w,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}

	var names []string

	t := r.find(req.URL.Path[1:], &names)
	if t == nil || t.rt == nil {
		http.NotFound(w, req)
		return
	}

	if h := t.rt.vb[v]; h != nil {
		h(w, req, names)
		return
	}

	http.Error(w,
		http.StatusText(http.StatusMethodNotAllowed),
		http.StatusMethodNotAllowed)
}

// Handle ...
func (r *router) Handle(verb Verb, path string, h Handler) {
	r.place(path[1:]).set(verb, h)
}

func (r *router) HandleAll(path string, h Handler) {
  n := r.place(path[1:])
  for i := 0; i < int(unknownVerb); i++ {
    n.set(Verb(i), h)
  }
}

func (r *router) Build() http.Handler {
	n := *r
	*r = router{}
	return &n
}

// New ...
func New() Builder {
	return &router{}
}
