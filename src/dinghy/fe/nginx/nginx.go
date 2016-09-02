package nginx

import (
	"os"
	"os/exec"
	"syscall"

	"dinghy/store"
)

// DefaultOptions ...
var DefaultOptions = Options{
	Command:   "nginx",
	ConfigDir: "/etc/nginx/conf.d",
}

// Service ...
type Service struct {
	p *os.Process
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

// Update ...
func (s *Service) Update(rts []*store.Route) error {
	// TODO(knorton): Write all the configs
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
	}, nil
}
