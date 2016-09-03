package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"dinghy/client/docker"
	"dinghy/client/proxy"
	"dinghy/client/routes"
)

func usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: %s [user@]addr docker_args...",
		filepath.Base(os.Args[0]))
	os.Exit(1)
}

func run(addr net.Addr, args []string) {
	// routes create --port=80 name host1 host2
	// routes ls
	// routes rm name
	// backends name set upstrea1 upstream2
	// backends name ls

	log.Println(args)
	if routes.CanRun(args) {
		routes.Run(addr, args)
	} else {
		docker.Run(addr, args)
	}
}

func main() {
	args := os.Args[1:]

	if len(args) < 2 {
		usage()
	}

	var addr proxy.Addr
	if err := addr.Parse(args[0]); err != nil {
		log.Panic(err)
	}

	p, err := proxy.Connect(&addr)
	if err != nil {
		log.Panic(err)
	}
	defer p.Close()

	run(p.Addr, args[1:])
}
