package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Addr ...
type Addr struct {
	User string
	Addr string
}

func parseAddr(a string, addr *Addr) error {
	ix := strings.Index(a, "@")
	if ix < 0 {
		u, err := user.Current()
		if err != nil {
			return err
		}

		addr.User = u.Username
	} else {
		addr.User = a[:ix]
		a = a[ix+1:]
	}

	addr.Addr = a
	if !strings.Contains(a, ":") {
		addr.Addr += ":22"
	}

	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: %s [user@]addr docker_args...",
		filepath.Base(os.Args[0]))
	os.Exit(1)
}

func getAgent() (ssh.AuthMethod, error) {
	addr := os.Getenv("SSH_AUTH_SOCK")
	if addr == "" {
		return nil, errors.New("no agent found")
	}

	c, err := net.Dial("unix", addr)
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeysCallback(agent.NewClient(c).Signers), nil
}

func serve(c *ssh.Client, n net.Conn) {
	defer n.Close()

	cc, err := c.Dial("tcp", "127.0.0.1:6660")
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		if _, err := io.Copy(n, cc); err != nil {
			log.Println(err)
			n.Close()
		}
	}()

	if _, err := io.Copy(cc, n); err != nil {
		log.Panic(err)
	}
}

func runServer(c *ssh.Client) (net.Addr, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			n, err := l.Accept()
			if err != nil {
				log.Panic(err)
			}

			go serve(c, n)
		}
	}()

	return l.Addr(), nil
}

func runDocker(host string, args []string) {
	pargs := []string{
		"-H",
		host,
	}
	for _, arg := range args {
		pargs = append(pargs, arg)
	}

	cmd := exec.Command("docker", pargs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			if ws, ok := err.ProcessState.Sys().(syscall.WaitStatus); ok {
				os.Exit(ws.ExitStatus())
			}
		}
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		usage()
	}

	var addr Addr
	if err := parseAddr(flag.Arg(0), &addr); err != nil {
		log.Panic(err)
	}

	auth, err := getAgent()
	if err != nil {
		log.Panic(err)
	}

	s, err := ssh.Dial("tcp", addr.Addr, &ssh.ClientConfig{
		User: addr.User,
		Auth: []ssh.AuthMethod{auth},
	})
	if err != nil {
		log.Panic(err)
	}
	defer s.Close()

	laddr, err := runServer(s)
	if err != nil {
		log.Panic(err)
	}

	// TODO(knorton): Handle route commands

	runDocker(laddr.String(), flag.Args()[1:])
}
