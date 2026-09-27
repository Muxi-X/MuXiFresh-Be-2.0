package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func TestCatchAllDispatchesByMethod(t *testing.T) {
	r := New()
	if err := r.Handle(http.MethodGet, "/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("get"))
	})); err != nil {
		t.Fatalf("Handle() = %v, want nil", err)
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v2/auth/login", nil))
	if rec.Body.String() != "get" {
		t.Fatalf("GET body = %q, want %q", rec.Body.String(), "get")
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v2/auth/login", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unregistered method status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCatchAllUsesNotFoundHandler(t *testing.T) {
	r := New()
	r.SetNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/anything", nil))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}
}

func TestCatchAllUsesNotAllowedHandler(t *testing.T) {
	r := New()
	r.SetNotAllowedHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/anything", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestCatchAllImplementsRouter(t *testing.T) {
	var _ httpx.Router = New()
}
