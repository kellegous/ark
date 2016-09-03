package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"dinghy/store"
)

const (
	routesCmd   = "routes"
	backendsCmd = "backends"
)

var errNotImplemented = errors.New("not implemented")

// CanRun ...
func CanRun(args []string) bool {
	return args[0] == routesCmd || args[0] == backendsCmd
}

// Run ...
func Run(laddr net.Addr, args []string) {
	switch args[0] {
	case routesCmd:
		runRoutes(laddr, args)
	case backendsCmd:
		runBackends(laddr, args)
	default:
		fmt.Fprintf(os.Stderr, "'%s' is not a command", args[1])
		os.Exit(1)
	}
}

func routesUsage() {
	fmt.Fprintln(os.Stderr, "routes usage")
	os.Exit(1)
}

func exit(err error) {
	if err != nil {
		log.Panic(err)
	}
	os.Exit(0)
}

func urlFor(laddr net.Addr, uri string) string {
	return fmt.Sprintf("http://%s%s", laddr.String(), uri)
}

func getJSON(laddr net.Addr, uri string, dst interface{}) error {
	res, err := http.Get(urlFor(laddr, uri))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("status: %d", res.StatusCode)
	}

	return json.NewDecoder(res.Body).Decode(dst)
}

func postJSON(laddr net.Addr, uri string, src, dst interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}

	res, err := http.Post(
		fmt.Sprintf("http://%s%s", laddr.String(), uri),
		"application/json",
		&buf)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("status: %d", res.StatusCode)
	}

	return json.NewDecoder(res.Body).Decode(dst)
}

func deleteRoute(laddr net.Addr, args []string) error {
	req, err := http.NewRequest(
		"DELETE",
		urlFor(laddr, fmt.Sprintf("/api/v1/routes/%s", args[0])),
		nil)
	if err != nil {
		return err
	}

	var c http.Client
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status: %d", res.StatusCode)
	}

	return nil
}

func createRoutes(laddr net.Addr, args []string) error {
	f := flag.NewFlagSet("create-routes", flag.PanicOnError)
	flagPort := f.Int("port", 80, "tcp port")
	f.Parse(args)

	if f.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "routes create help")
		os.Exit(1)
	}

	rt := store.Route{
		Name:  f.Arg(0),
		Port:  int32(*flagPort),
		Hosts: f.Args()[1:],
	}

	if err := postJSON(laddr, "/api/v1/routes", &rt, &rt); err != nil {
		return err
	}

	fmt.Println(rt.Name)
	return nil
}

func runRoutes(laddr net.Addr, args []string) {
	if len(args) < 2 {
		routesUsage()
	}

	switch args[1] {
	case "create":
		exit(createRoutes(laddr, args[2:]))
	case "rm":
		exit(deleteRoute(laddr, args[2:]))
	case "ls":
		exit(errNotImplemented)
	default:
		fmt.Fprintf(os.Stderr, "'%s' is not a routes command", args[1])
		os.Exit(1)
	}
}

func backendsUsage() {
	fmt.Fprintln(os.Stderr, "backends usage")
	os.Exit(1)
}

func setBackends(laddr net.Addr, name string, args []string) error {
	var bes []string
	if err := postJSON(
		laddr,
		fmt.Sprintf("/api/v1/routes/%s/backends", name),
		&args,
		&bes); err != nil {
		return err
	}

	for _, be := range bes {
		fmt.Println(be)
	}
	return nil
}

func getBackends(laddr net.Addr, name string) error {
	var bes []string
	if err := getJSON(
		laddr,
		fmt.Sprintf("/api/v1/routes/%s/backends", name),
		&bes); err != nil {
		return err
	}

	for _, be := range bes {
		fmt.Println(be)
	}
	return nil
}

func runBackends(laddr net.Addr, args []string) {
	if len(args) < 3 {
		backendsUsage()
	}

	switch args[2] {
	case "set":
		exit(setBackends(laddr, args[1], args[3:]))
	case "get":
		exit(getBackends(laddr, args[1]))
	}
	exit(errNotImplemented)
}
