package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"MuXiFresh-Be-2.0/app/gateway/internal/config"
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

func TestGatewayRoutesEveryPrefixToItsUpstream(t *testing.T) {
	names := []string{"auth", "user", "exam", "task", "review", "form", "schedule"}
	upstreams := make([]config.Upstream, 0, len(names))
	for _, name := range names {
		srv, u := startUpstream(t, name)
		defer srv.Close()
		upstreams = append(upstreams, u)
	}

	var c config.Config
	c.Name = "gateway-test"
	c.Host = "127.0.0.1"
	c.Upstreams = upstreams

	server, err := buildServer(c)
	if err != nil {
		t.Fatalf("buildServer() = %v, want nil", err)
	}
	defer server.Stop()

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
		server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if rec.Code != http.StatusOK || rec.Body.String() != tt.want {
			t.Errorf("%s -> (%d, %q), want (200, %q)", tt.path, rec.Code, rec.Body.String(), tt.want)
		}
	}
}

func TestBuildServerFailsFastOnUnknownUpstream(t *testing.T) {
	var c config.Config
	c.Name = "gateway-test"
	c.Upstreams = []config.Upstream{{Name: "auth", Host: "127.0.0.1", Port: 1}}

	if _, err := buildServer(c); err == nil {
		t.Fatal("buildServer() with missing upstreams = nil, want error")
	}
}
