package proxy

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/user"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Addr ...
type Addr struct {
	User string
	Addr string
}

// Proxy ...
type Proxy struct {
	Addr net.Addr
	c    *ssh.Client
}

// Parse ...
func (a *Addr) Parse(str string) error {
	ix := strings.Index(str, "@")
	if ix < 0 {
		u, err := user.Current()
		if err != nil {
			return err
		}

		a.User = u.Username
	} else {
		a.User = str[:ix]
		str = str[ix+1:]
	}

	a.Addr = str
	if !strings.Contains(str, ":") {
		a.Addr += ":22"
	}

	return nil
}

// Close ...
func (p *Proxy) Close() error {
	return p.c.Close()
}

// Connect ...
func Connect(addr *Addr) (*Proxy, error) {
	auth, err := getAgent()
	if err != nil {
		return nil, err
	}

	c, err := ssh.Dial("tcp", addr.Addr, &ssh.ClientConfig{
		User: addr.User,
		Auth: []ssh.AuthMethod{auth},
	})
	if err != nil {
		return nil, err
	}

	laddr, err := listen(c)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		Addr: laddr,
		c:    c,
	}, nil
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

func listen(c *ssh.Client) (net.Addr, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}

	go func() {
		prx := &httputil.ReverseProxy{
			Director: func(r *http.Request) {
				r.URL.Scheme = "http"
				r.URL.Host = "127.0.0.1"
			},
			Transport: &http.Transport{
				Dial: func(network, address string) (net.Conn, error) {
					return c.Dial("tcp", "127.0.0.1:6660")
				},
				DisableKeepAlives: true,
			},
		}

		srv := &http.Server{
			Handler: prx,
		}

		log.Panic(srv.Serve(l))
	}()

	return l.Addr(), nil
}
