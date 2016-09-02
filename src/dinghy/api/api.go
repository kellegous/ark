package api

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"dinghy/fe"
	"dinghy/store"
	"dinghy/web/router"
)

// Options ...
type Options struct {
	DockerDialer func() (net.Conn, error)
	LoadBalancer interface{}
}

// Context ...
type Context struct {
	Store        *store.Store
	LoadBalancer fe.Service
	DockerDialer func() (net.Conn, error)
}

func getRoutes(ctx *Context,
	w http.ResponseWriter,
	r *http.Request,
	names []string) {
	// TODO(knorton): Fetch the routes from the nginx configs. This will also
	// resolve the routes to docker containers.
}

func postRoutes(ctx *Context,
	w http.ResponseWriter,
	r *http.Request,
	names []string) {
	// TODO(knorton): Write a new route file to the nginx config.
}

func getRoute(ctx *Context,
	w http.ResponseWriter,
	r *http.Request,
	names []string) {
	// TODO(knorton): Read the route file from nginx config.
}

func proxyToDocker(w http.ResponseWriter, r *http.Request, ctx *Context) error {
	c, err := ctx.DockerDialer()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := r.Write(c); err != nil {
		return err
	}

	hj := w.(http.Hijacker)

	dc, bw, err := hj.Hijack()
	if err != nil {
		return err
	}
	defer dc.Close()

	if _, err := io.Copy(bw, c); err != nil {
		return err
	}

	return bw.Flush()
}

// ListenAndServe ...
func ListenAndServe(addr string, ctx *Context) error {
	r := router.New()

	r.Handle(router.Get, "/api/v1/routes",
		func(w http.ResponseWriter, r *http.Request, names []string) {
			getRoutes(ctx, w, r, names)
		})

	r.Handle(router.Post, "/api/v1/routes",
		func(w http.ResponseWriter, r *http.Request, names []string) {
			postRoutes(ctx, w, r, names)
		})

	r.Handle(router.Get, "/api/v1/routes/*",
		func(w http.ResponseWriter, r *http.Request, names []string) {
			getRoute(ctx, w, r, names)
		})

	h := r.Build()

	return http.ListenAndServe(addr, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				h.ServeHTTP(w, r)
				return
			}

			if err := proxyToDocker(w, r, ctx); err != nil {
				log.Panic(err)
			}
		}))
}
