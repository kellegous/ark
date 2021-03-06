package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"golang.org/x/net/context"

	"ark/docker"
	"ark/fe"
	"ark/store"
	"ark/web/router"
)

// Options ...
type Options struct {
	DockerDialer func() (net.Conn, error)
	LoadBalancer interface{}
}

// Context ...
type Context struct {
	Store        store.Store
	LoadBalancer fe.Service
	DockerDialer func() (net.Conn, error)
}

func (c *Context) update() error {
	rts, err := c.Store.LoadAll()
	if err != nil {
		return err
	}

	return c.LoadBalancer.Update(rts)
}

func emitJSONError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	}); err != nil {
		log.Panic(err)
	}
}

func emitJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Panic(err)
	}
}

func emitNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func getRoutes(ctx *Context,
	w http.ResponseWriter,
	r *http.Request,
	names []string) {

	rts, err := ctx.Store.LoadAll()
	if err != nil {
		emitJSONError(w, err, http.StatusInternalServerError)
		return
	}

	emitJSON(w, rts)
}

func validateRoute(r *store.Route) error {
	if r.Name == "" {
		return errors.New("name is required")
	}

	if r.Port == 0 {
		return errors.New("port is required")
	}

	if len(r.Hosts) == 0 {
		return errors.New("at least one host is required")
	}

	return nil
}

func postRoutes(ctx *Context,
	w http.ResponseWriter,
	r *http.Request,
	names []string) {

	var rt store.Route

	if err := json.NewDecoder(r.Body).Decode(&rt); err != nil {
		emitJSONError(w, err, http.StatusBadRequest)
		return
	}

	if err := validateRoute(&rt); err != nil {
		emitJSONError(w, err, http.StatusBadRequest)
		return
	}

	if err := ctx.Store.Save(&rt); err != nil {
		emitJSONError(w, err, http.StatusInternalServerError)
		return
	}

	if err := ctx.update(); err != nil {
		emitJSONError(w, err, http.StatusInternalServerError)
	}

	emitJSON(w, &rt)
}

func getRoute(ctx *Context,
	w http.ResponseWriter,
	r *http.Request,
	names []string) {
	var rt store.Route
	err := ctx.Store.Load(names[0], &rt)
	if err == store.ErrNotFound {
		emitJSONError(w, fmt.Errorf("%s not found", names[0]), http.StatusNotFound)
		return
	}
	emitJSON(w, &rt)
}

func delRoute(ctx *Context,
	w http.ResponseWriter,
	r *http.Request,
	names []string) {

	err := ctx.Store.Delete(names[0])
	if err == store.ErrNotFound {
		emitJSONError(w, err, http.StatusNotFound)
		return
	} else if err != nil {
		emitJSONError(w, err, http.StatusInternalServerError)
		return
	}

	if err := ctx.update(); err != nil {
		emitJSONError(w, err, http.StatusInternalServerError)
		return
	}

	emitNoContent(w)
}

func getBackends(ctx *Context,
	w http.ResponseWriter,
	r *http.Request,
	names []string) {
	// Lookup the route
	var rt store.Route
	err := ctx.Store.Load(names[0], &rt)
	if err == store.ErrNotFound {
		emitJSONError(w, fmt.Errorf("%s not found", names[0]), http.StatusNotFound)
		return
	}

	ips, err := docker.ParseRefs(rt.Backends)
	if err != nil {
		emitJSONError(w, err, http.StatusInternalServerError)
		return
	}

	// Translate from ip address to container id
	cids, err := docker.ToContainers(context.Background(), ips)
	if err != nil {
		emitJSONError(w, err, http.StatusInternalServerError)
		return
	}

	emitJSON(w, cids)
}

func readContainerRefs(ctx context.Context, r io.Reader) ([]string, error) {
	var bes []string
	if err := json.NewDecoder(r).Decode(&bes); err != nil {
		return nil, err
	}

	cids, err := docker.ParseRefs(bes)
	if err != nil {
		return nil, err
	}

	ips, err := docker.ToIPAddresses(ctx, cids)
	if err != nil {
		return nil, err
	}

	for i, ip := range ips {
		bes[i] = ip.String()
	}

	return bes, nil
}

func postBackends(ctx *Context,
	w http.ResponseWriter,
	r *http.Request,
	names []string) {

	bes, err := readContainerRefs(
		context.Background(),
		r.Body)
	if docker.IsNotFound(err) {
		emitJSONError(w, err, http.StatusNotFound)
		return
	} else if err != nil {
		emitJSONError(w, err, http.StatusBadRequest)
		return
	}

	var rt store.Route
	err = ctx.Store.Load(names[0], &rt)
	if err == store.ErrNotFound {
		emitJSONError(w, fmt.Errorf("route not found: '%s'", names[0]), http.StatusNotFound)
		return
	} else if err != nil {
		emitJSONError(w, err, http.StatusInternalServerError)
		return
	}

	rt.Backends = bes
	if err := ctx.Store.Save(&rt); err != nil {
		emitJSONError(w, err, http.StatusInternalServerError)
		return
	}

	if err := ctx.update(); err != nil {
		emitJSONError(w, err, http.StatusInternalServerError)
		return
	}

	emitJSON(w, rt.Backends)
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

// Handler ...
func Handler(ctx *Context) http.Handler {
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

	r.Handle(router.Delete, "/api/v1/routes/*",
		func(w http.ResponseWriter, r *http.Request, names []string) {
			delRoute(ctx, w, r, names)
		})

	r.Handle(router.Get, "/api/v1/routes/*/backends",
		func(w http.ResponseWriter, r *http.Request, names []string) {
			getBackends(ctx, w, r, names)
		})

	r.Handle(router.Post, "/api/v1/routes/*/backends",
		func(w http.ResponseWriter, r *http.Request, names []string) {
			postBackends(ctx, w, r, names)
		})

	return r.Build()
}

// ListenAndServe ...
func ListenAndServe(addr string, ctx *Context) error {

	h := Handler(ctx)

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
