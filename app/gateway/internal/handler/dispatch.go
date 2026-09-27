// Package handler dispatches gateway requests to an upstream by longest
// matching path prefix.
package handler

import (
	"net/http"

	"MuXiFresh-Be-2.0/app/gateway/internal/proxy"
	"MuXiFresh-Be-2.0/app/gateway/internal/route"
)

const healthPath = "/health"

// Dispatcher routes requests to the configured upstreams.
type Dispatcher struct {
	rules   []route.Rule
	proxies *proxy.Set
}

// NewDispatcher returns a dispatcher for the given rules and proxies.
func NewDispatcher(rules []route.Rule, proxies *proxy.Set) *Dispatcher {
	return &Dispatcher{rules: rules, proxies: proxies}
}

// ServeHTTP implements http.Handler.
func (d *Dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == healthPath {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	}

	upstream, ok := route.Match(d.rules, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	p, ok := d.proxies.Get(upstream)
	if !ok {
		http.Error(w, "upstream not configured", http.StatusBadGateway)
		return
	}

	p.ServeHTTP(w, r)
}
