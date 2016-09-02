package main

import (
	"flag"
	"log"
	"net"

	"dinghy/api"
	"dinghy/fe/nginx"
	"dinghy/store"
)

func main() {
	flagAddr := flag.String("addr", ":6660", "")
	flagSock := flag.String("sock", "/var/run/docker.sock", "")
	flagStore := flag.String("data", "routes.db", "")
	flag.Parse()

	db, err := store.Open(*flagStore)
	if err != nil {
		log.Panic(err)
	}

	fe, err := nginx.Start(&nginx.Options{
		Command:   "nginx",
		ConfigDir: "/tmp/foo",
	})
	if err != nil {
		log.Panic(err)
	}

	if err := fe.Update([]*store.Route{
		&store.Route{
			Name:     "kunk",
			Port:     80,
			Hosts:    []string{"a.kellegous.com"},
			Backends: []string{"127.0.0.1:6070"},
		},
	}); err != nil {
		log.Panic(err)
	}

	ctx := api.Context{
		Store:        db,
		LoadBalancer: fe,
		DockerDialer: func() (net.Conn, error) {
			return net.Dial("unix", *flagSock)
		},
	}

	log.Panic(api.ListenAndServe(*flagAddr, &ctx))
}
