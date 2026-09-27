// Package proxy builds one reverse proxy per configured upstream.
package proxy

import (
	"fmt"
	"net"
	"net/http/httputil"
	"net/url"
	"strconv"

	"MuXiFresh-Be-2.0/app/gateway/internal/config"
)

// Set holds one reverse proxy per upstream name.
type Set struct {
	proxies map[string]*httputil.ReverseProxy
}

// NewSet builds a reverse proxy for each upstream. The proxies rewrite only
// scheme and host, so method, path, query, body and headers reach the upstream
// unchanged.
func NewSet(upstreams []config.Upstream) (*Set, error) {
	set := &Set{proxies: make(map[string]*httputil.ReverseProxy, len(upstreams))}
	for _, upstream := range upstreams {
		target, err := url.Parse("http://" + net.JoinHostPort(upstream.Host, strconv.Itoa(upstream.Port)))
		if err != nil {
			return nil, fmt.Errorf("upstream %q: %w", upstream.Name, err)
		}
		set.proxies[upstream.Name] = httputil.NewSingleHostReverseProxy(target)
	}
	return set, nil
}

// Get returns the reverse proxy for the given upstream name.
func (s *Set) Get(name string) (*httputil.ReverseProxy, bool) {
	p, ok := s.proxies[name]
	return p, ok
}
