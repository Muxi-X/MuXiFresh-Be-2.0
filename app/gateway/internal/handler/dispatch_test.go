package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"MuXiFresh-Be-2.0/app/gateway/internal/config"
	"MuXiFresh-Be-2.0/app/gateway/internal/proxy"
	"MuXiFresh-Be-2.0/app/gateway/internal/route"
)

func startUpstream(t *testing.T, name string) (*httptest.Server, config.Upstream) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(name))
	}))

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}

	return srv, config.Upstream{Name: name, Host: u.Hostname(), Port: port}
}

func newDispatcher(t *testing.T, upstreams []config.Upstream) *Dispatcher {
	t.Helper()

	proxies, err := proxy.NewSet(upstreams)
	if err != nil {
		t.Fatal(err)
	}

	return NewDispatcher(route.Default(), proxies)
}

func TestDispatcherRoutesByPrefix(t *testing.T) {
	names := []string{"auth", "user", "exam", "task", "review", "form", "schedule"}
	upstreams := make([]config.Upstream, 0, len(names))
	for _, name := range names {
		srv, u := startUpstream(t, name)
		defer srv.Close()
		upstreams = append(upstreams, u)
	}

	d := newDispatcher(t, upstreams)

	tests := []struct {
		path string
		want string
	}{
		{path: "/api/v2/auth/login", want: "auth"},
		{path: "/api/v2/users/info/1", want: "user"},
		{path: "/api/v2/user/test/result", want: "exam"},
		{path: "/api/v2/task/assigned/list", want: "task"},
		{path: "/api/v2/review/export", want: "review"},
		{path: "/api/v2/form/view", want: "form"},
		{path: "/api/v2/recruit/deadline", want: "form"},
		{path: "/api/v2/schedule/create", want: "schedule"},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		d.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if rec.Code != http.StatusOK || rec.Body.String() != tt.want {
			t.Errorf("%s -> (%d, %q), want (200, %q)", tt.path, rec.Code, rec.Body.String(), tt.want)
		}
	}
}

func TestDispatcherReturnsNotFoundForUnknownPrefix(t *testing.T) {
	srv, u := startUpstream(t, "auth")
	defer srv.Close()

	d := newDispatcher(t, []config.Upstream{u})

	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v2/intro", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDispatcherHealth(t *testing.T) {
	d := newDispatcher(t, nil)

	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "ok")
	}
}

func TestDispatcherPassesRequestAndResponseThrough(t *testing.T) {
	var got struct {
		method string
		path   string
		query  string
		body   string
		auth   string
		ctype  string
		host   string
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got.method = r.Method
		got.path = r.URL.Path
		got.query = r.URL.RawQuery
		got.body = string(body)
		got.auth = r.Header.Get("Authorization")
		got.ctype = r.Header.Get("Content-Type")
		got.host = r.Host

		w.Header().Set("X-Upstream", "yes")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
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

	d := newDispatcher(t, []config.Upstream{{Name: "review", Host: u.Hostname(), Port: port}})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "http://fresh-be.muxistudio.xyz/api/v2/review?group=All", strings.NewReader("payload"))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")
	d.ServeHTTP(rec, req)

	if got.method != http.MethodPost {
		t.Errorf("upstream method = %q, want %q", got.method, http.MethodPost)
	}
	if got.path != "/api/v2/review" {
		t.Errorf("upstream path = %q, want %q", got.path, "/api/v2/review")
	}
	if got.query != "group=All" {
		t.Errorf("upstream query = %q, want %q", got.query, "group=All")
	}
	if got.body != "payload" {
		t.Errorf("upstream body = %q, want %q", got.body, "payload")
	}
	if got.auth != "Bearer token" {
		t.Errorf("upstream Authorization = %q, want %q", got.auth, "Bearer token")
	}
	if got.ctype != "application/json" {
		t.Errorf("upstream Content-Type = %q, want %q", got.ctype, "application/json")
	}
	if got.host != "fresh-be.muxistudio.xyz" {
		t.Errorf("upstream Host = %q, want %q", got.host, "fresh-be.muxistudio.xyz")
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != "created" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "created")
	}
	if rec.Header().Get("X-Upstream") != "yes" {
		t.Errorf("response header X-Upstream = %q, want %q", rec.Header().Get("X-Upstream"), "yes")
	}
}
