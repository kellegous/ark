package nginx

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"

	"dinghy/store"
)

// DefaultOptions ...
var DefaultOptions = Options{
	Command:   "nginx",
	ConfigDir: "/etc/nginx/conf.d",
}

const tpl = `
server {
  listen {{.Port}};
  root /var/www/html;
  index index.html;

  server_name {{.ServerName}};

  location / {
    proxy_pass_header Server;
    proxy_set_header Host $http_host;
    proxy_redirect off;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Scheme $scheme;
    proxy_pass http://be{{.ID}};
  }
}

upstream be{{.ID}} {
  {{range .Backends}}
  server {{.}};
  {{end}}
}
`

// Service ...
type Service struct {
	p *os.Process
	o *Options
}

// Options ...
type Options struct {
	Command   string
	ConfigDir string
}

// Reload ...
func (s *Service) Reload() error {
	return s.p.Signal(syscall.SIGHUP)
}

func nameFor(r *store.Route) string {
	h := sha1.New()
	h.Write([]byte(r.Name))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func writeTo(dir string, r *store.Route) error {
	id := nameFor(r)

	dst := filepath.Join(dir, fmt.Sprintf("%s.conf", id))

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer w.Close()

	t, err := template.New("tpl").Parse(tpl)
	if err != nil {
		return err
	}

	data := struct {
		*store.Route
		ID         string
		ServerName string
	}{
		r,
		id,
		strings.Join(r.Hosts, " "),
	}

	return t.Execute(w, &data)
}

// Update ...
func (s *Service) Update(rts []*store.Route) error {
	files, err := filepath.Glob(filepath.Join(s.o.ConfigDir, "*.conf"))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}

	for _, rt := range rts {
		if len(rt.Backends) == 0 {
			continue
		}

		if err := writeTo(s.o.ConfigDir, rt); err != nil {
			return err
		}
	}
	return s.Reload()
}

// Start ...
func Start(opts *Options) (*Service, error) {
	if opts == nil {
		opts = &DefaultOptions
	}

	cmd := exec.Command(opts.Command, "-g", "daemon off;")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Service{
		p: cmd.Process,
		o: opts,
	}, nil
}
