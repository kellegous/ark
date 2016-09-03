package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

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
	// TODO(knorton): Fix this.
	fmt.Fprintln(os.Stderr, "routes usage")
	os.Exit(1)
}

// urlFor produces a URL from the address and the uri
func urlFor(laddr net.Addr, uri string) string {
	return fmt.Sprintf("http://%s%s", laddr.String(), uri)
}

func decodeJSON(res *http.Response, data interface{}) error {
	switch res.StatusCode {
	case http.StatusOK:
		return json.NewDecoder(res.Body).Decode(data)
	case http.StatusNoContent:
		return nil
	}

	var e struct {
		Error string `json:"error"`
	}

	if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
		return fmt.Errorf("%d: %s", res.StatusCode,
			http.StatusText(res.StatusCode))
	}

	return errors.New(e.Error)
}

func getJSON(laddr net.Addr, uri string, dst interface{}) error {
	res, err := http.Get(urlFor(laddr, uri))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return decodeJSON(res, dst)
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

	return decodeJSON(res, dst)
}

func errorLn(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func deleteRoute(laddr net.Addr, args []string) {
	req, err := http.NewRequest(
		"DELETE",
		urlFor(laddr, fmt.Sprintf("/api/v1/routes/%s", args[0])),
		nil)
	if err != nil {
		errorLn(err.Error())
	}

	var c http.Client
	res, err := c.Do(req)
	if err != nil {
		errorLn(err.Error())
	}
	defer res.Body.Close()

	if err := decodeJSON(res, nil); err != nil {
		errorLn(err.Error())
	}
}

func createRoutes(laddr net.Addr, args []string) {
	f := flag.NewFlagSet("create-routes", flag.PanicOnError)
	flagPort := f.Int("port", 80, "tcp port")
	f.Parse(args)

	if f.NArg() < 2 {
		errorLn("routes create help")
	}

	rt := store.Route{
		Name:  f.Arg(0),
		Port:  int32(*flagPort),
		Hosts: f.Args()[1:],
	}

	if err := postJSON(laddr, "/api/v1/routes", &rt, &rt); err != nil {
		errorLn(err.Error())
	}

	fmt.Println(rt.Name)
}

func listRoutes(laddr net.Addr, args []string) {
	var rts []*store.Route
	if err := getJSON(laddr, "/api/v1/routes", &rts); err != nil {
		errorLn(err.Error())
	}

	fmt.Printf("%- 15s % 5s  %- 30s %-30s\n", "NAME", "PORT", "HOSTS", "BACKENDS")
	for _, rt := range rts {
		fmt.Printf("%- 15s % 5d  %- 30s %- 30s\n",
			rt.Name,
			rt.Port,
			strings.Join(rt.Hosts, ","),
			strings.Join(rt.Backends, ","))
	}
}

func runRoutes(laddr net.Addr, args []string) {
	if len(args) < 2 {
		routesUsage()
	}

	switch args[1] {
	case "create":
		createRoutes(laddr, args[2:])
	case "rm":
		deleteRoute(laddr, args[2:])
	case "ls":
		listRoutes(laddr, args[2:])
	default:
		errorf("'%s' is not a routes command.\n", args[1])
	}

}

func backendsUsage() {
	fmt.Fprintln(os.Stderr, "backends usage")
	os.Exit(1)
}

func setBackends(laddr net.Addr, name string, args []string) {
	var bes []string
	if err := postJSON(
		laddr,
		fmt.Sprintf("/api/v1/routes/%s/backends", name),
		&args,
		&bes); err != nil {
		errorLn(err.Error())
	}

	for _, be := range bes {
		fmt.Println(be)
	}
}

func getBackends(laddr net.Addr, name string) {
	var bes []string
	if err := getJSON(
		laddr,
		fmt.Sprintf("/api/v1/routes/%s/backends", name),
		&bes); err != nil {
		errorLn(err.Error())
	}

	for _, be := range bes {
		fmt.Println(be)
	}
}

func runBackends(laddr net.Addr, args []string) {
	if len(args) < 3 {
		backendsUsage()
	}

	switch args[2] {
	case "set":
		setBackends(laddr, args[1], args[3:])
	case "get":
		getBackends(laddr, args[1])
	default:
		errorf("'%s' is not a backends command.\n", args[2])
	}
}
