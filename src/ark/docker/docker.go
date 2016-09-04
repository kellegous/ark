package docker

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
)

const (
	defaultHost = "unix:///var/run/docker.sock"
)

type errNotFound string

// Ref ...
type Ref struct {
	Addr string
	Port int
}

func (e errNotFound) Error() string {
	return string(e)
}

// IsNotFound ...
func IsNotFound(err error) bool {
	_, ok := err.(errNotFound)
	return ok
}

// MarshalJSON ...
func (r *Ref) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *Ref) String() string {
	return fmt.Sprintf("%s:%d", r.Addr, r.Port)
}

// ParseRef ...
func ParseRef(str string) (*Ref, error) {
	p := strings.SplitN(str, ":", 2)
	if len(p) != 2 {
		return nil, fmt.Errorf("malformed address: %s", str)
	}

	pt, err := strconv.ParseUint(p[1], 10, 32)
	if err != nil {
		return nil, err
	}

	return &Ref{
		Addr: p[0],
		Port: int(pt),
	}, nil
}

// ParseRefs ...
func ParseRefs(strs []string) ([]*Ref, error) {
	refs := make([]*Ref, 0, len(strs))
	for _, str := range strs {
		ref, err := ParseRef(str)
		if err != nil {
			return nil, err
		}

		refs = append(refs, ref)
	}
	return refs, nil
}

func newClient() (*client.Client, error) {
	return client.NewClient(
		defaultHost,
		client.DefaultVersion,
		nil,
		map[string]string{"User-Agent": "Ark"})
}

func fetchContainerMap(
	ctx context.Context,
	tx func(*types.Container) (string, string)) (map[string]string, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}

	ctrs, err := c.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	idx := map[string]string{}

	for _, ctr := range ctrs {
		k, v := tx(&ctr)
		idx[k] = v
	}

	return idx, nil
}

func transform(
	ctx context.Context,
	refs []*Ref,
	tx func(c *types.Container) (string, string)) ([]*Ref, error) {
	res := make([]*Ref, 0, len(refs))

	idx, err := fetchContainerMap(ctx, tx)
	if err != nil {
		return nil, err
	}

	for _, ref := range refs {
		addr := idx[ref.Addr]
		if addr == "" {
			return nil, fmt.Errorf("not found: %s", ref.Addr)
		}

		res = append(res, &Ref{
			Addr: addr,
			Port: ref.Port,
		})
	}

	return res, nil
}

// ToContainers ...
func ToContainers(ctx context.Context, refs []*Ref) ([]*Ref, error) {
	return transform(ctx, refs, func(c *types.Container) (string, string) {
		return c.NetworkSettings.Networks["bridge"].IPAddress, c.ID[:12]
	})
}

// ToIPAddresses ...
func ToIPAddresses(ctx context.Context, refs []*Ref) ([]*Ref, error) {
	return transform(ctx, refs, func(c *types.Container) (string, string) {
		return c.ID[:12], c.NetworkSettings.Networks["bridge"].IPAddress
	})
}
