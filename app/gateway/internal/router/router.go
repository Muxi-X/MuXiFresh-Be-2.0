// Package router provides an httpx.Router that dispatches by HTTP method only.
//
// go-zero's built-in patRouter cannot match multi-segment catch-all paths, so
// the gateway registers every method under one catch-all path and lets the
// dispatcher match the real prefix. This router preserves the per-method
// middleware chain that rest.Server built.
package router

import "net/http"

// CatchAll routes a request to the handler registered for its HTTP method,
// ignoring the path.
type CatchAll struct {
	handlers   map[string]http.Handler
	notFound   http.Handler
	notAllowed http.Handler
}

// New returns an empty CatchAll router.
func New() *CatchAll {
	return &CatchAll{handlers: make(map[string]http.Handler)}
}

// Handle stores the per-method handler. The path is ignored because matching
// happens in the business dispatcher.
func (r *CatchAll) Handle(method, path string, handler http.Handler) error {
	r.handlers[method] = handler
	return nil
}

// SetNotFoundHandler sets the fallback used when a method is unregistered and
// no not-allowed handler exists.
func (r *CatchAll) SetNotFoundHandler(handler http.Handler) {
	r.notFound = handler
}

// SetNotAllowedHandler sets the fallback for unregistered methods.
func (r *CatchAll) SetNotAllowedHandler(handler http.Handler) {
	r.notAllowed = handler
}

// ServeHTTP dispatches to the handler registered for req.Method.
func (r *CatchAll) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if handler, ok := r.handlers[req.Method]; ok {
		handler.ServeHTTP(w, req)
		return
	}
	if r.notAllowed != nil {
		r.notAllowed.ServeHTTP(w, req)
		return
	}
	if r.notFound != nil {
		r.notFound.ServeHTTP(w, req)
		return
	}
	http.NotFound(w, req)
}
