package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"MuXiFresh-Be-2.0/app/gateway/internal/config"
)

func TestSetForwardsToUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Seen-Path", r.URL.Path)
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("up"))
	}))
	defer upstream.Close()

	u, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}

	set, err := NewSet([]config.Upstream{{Name: "auth", Host: u.Hostname(), Port: port}})
	if err != nil {
		t.Fatalf("NewSet() = %v, want nil", err)
	}

	p, ok := set.Get("auth")
	if !ok {
		t.Fatal("Get(auth) = not found, want proxy")
	}
	if _, ok := set.Get("missing"); ok {
		t.Fatal("Get(missing) = found, want not found")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/v2/auth/login?x=1", nil)
	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}
	if rec.Body.String() != "up" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "up")
	}
	if got := rec.Header().Get("X-Seen-Path"); got != "/api/v2/auth/login" {
		t.Fatalf("upstream path = %q, want %q", got, "/api/v2/auth/login")
	}
}

func TestProxyReturnsBadGatewayWhenUpstreamDown(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	u, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	upstream.Close()

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}

	set, err := NewSet([]config.Upstream{{Name: "auth", Host: u.Hostname(), Port: port}})
	if err != nil {
		t.Fatalf("NewSet() = %v, want nil", err)
	}

	p, ok := set.Get("auth")
	if !ok {
		t.Fatal("Get(auth) = not found, want proxy")
	}

	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://example.com/api/v2/auth/login", nil))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
}
