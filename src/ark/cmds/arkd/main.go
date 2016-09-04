package main

import (
	"flag"
	"log"
	"net"

	"ark/api"
	"ark/fe/nginx"
	"ark/store"
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

	fe, err := nginx.Start(nil)
	if err != nil {
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
