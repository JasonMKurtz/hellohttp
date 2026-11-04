package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/jasonmkurtz/hellohttp/src/lib/db"
	routetypes "github.com/jasonmkurtz/hellohttp/src/lib/routes"
)

func TestVersionReportsStampedVersion(t *testing.T) {
	version = "abc123"

	rec := httptest.NewRecorder()
	Version(rec, httptest.NewRequest("GET", "/version", nil), routetypes.Route{})

	if got := rec.Body.String(); got != "abc123" {
		t.Errorf("body %q, want %q", got, "abc123")
	}
}

// Greet must still answer when MySQL and hellohttp-backend aren't running,
// instead of panicking (which the load balancer turns into a 502).
func TestGreetWithoutDependencies(t *testing.T) {
	// .invalid hostnames are guaranteed never to resolve.
	app = routetypes.Routes{}
	app.AddService("backend.invalid", 80)
	app.Services["hellohttp-backend"] = app.Services["backend.invalid"]
	app.AddDatabase(db.Database{Host: "mysql.invalid", Port: "3306", User: "root", Db: "hello"})

	route := routetypes.Route{Route: "^/greet/?(?P<name>[a-zA-Z]+)?/?$"}

	cases := map[string]string{
		"/greet":       "Hey guest!\n",
		"/greet/Jason": "Hey Jason!\n",
	}
	for path, want := range cases {
		rec := httptest.NewRecorder()
		Greet(rec, httptest.NewRequest("GET", path, nil), route)

		if rec.Code != 200 {
			t.Errorf("%s: status %d, want 200", path, rec.Code)
		}
		if got := rec.Body.String(); got != want {
			t.Errorf("%s: body %q, want %q", path, got, want)
		}
	}
}

func TestIPFamily(t *testing.T) {
	cases := []struct {
		name string
		addr string
		want string
	}{
		{name: "ipv4 host:port", addr: "10.4.0.5:5000", want: "IPv4"},
		{name: "ipv6 host:port", addr: "[2600:1900:4000:1249:0:1:0:5]:5000", want: "IPv6"},
		{name: "bare ipv6", addr: "2600:1900:4000:1249::1", want: "IPv6"},
		{name: "ipv4-mapped ipv6", addr: "[::ffff:192.0.2.1]:1234", want: "IPv4"},
		{name: "no address", addr: "", want: ""},
		{name: "hostname", addr: "hellohttp:80", want: ""},
	}
	for _, c := range cases {
		if got := ipFamily(hostIP(c.addr)); got != c.want {
			t.Errorf("%s: ipFamily(hostIP(%q)) = %q, want %q", c.name, c.addr, got, c.want)
		}
	}
}

func TestWhoAmIReportsConnectionFamilies(t *testing.T) {
	cases := []struct {
		name             string
		remote           string
		local            net.Addr
		wantClientIP     string
		wantClientFamily string
		wantServerIP     string
		wantServerFamily string
	}{
		{
			name:             "ipv6 connection",
			remote:           "[2600:1900:4000:1249:0:1::]:41234",
			local:            &net.TCPAddr{IP: net.ParseIP("2600:1900:4000:1249:0:1:0:5"), Port: 5000},
			wantClientIP:     "2600:1900:4000:1249:0:1::",
			wantClientFamily: "IPv6",
			wantServerIP:     "2600:1900:4000:1249:0:1:0:5",
			wantServerFamily: "IPv6",
		},
		{
			name:             "ipv4 connection",
			remote:           "10.0.0.9:41234",
			local:            &net.TCPAddr{IP: net.ParseIP("10.4.0.5"), Port: 5000},
			wantClientIP:     "10.0.0.9",
			wantClientFamily: "IPv4",
			wantServerIP:     "10.4.0.5",
			wantServerFamily: "IPv4",
		},
		{
			name:             "no local address",
			remote:           "10.0.0.9:41234",
			local:            nil,
			wantClientIP:     "10.0.0.9",
			wantClientFamily: "IPv4",
			wantServerIP:     "",
			wantServerFamily: "",
		},
	}
	for _, c := range cases {
		req := httptest.NewRequest("GET", "/whoami", nil)
		req.RemoteAddr = c.remote
		if c.local != nil {
			req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, c.local))
		}

		rec := httptest.NewRecorder()
		WhoAmI(rec, req, routetypes.Route{})

		var got map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("%s: body %q isn't JSON: %v", c.name, rec.Body.String(), err)
		}

		want := map[string]string{
			"ClientIP":     c.wantClientIP,
			"ClientFamily": c.wantClientFamily,
			"ServerIP":     c.wantServerIP,
			"ServerFamily": c.wantServerFamily,
		}
		for field, w := range want {
			if got[field] != w {
				t.Errorf("%s: %s = %v, want %q", c.name, field, got[field], w)
			}
		}
		if _, ok := got["PodIPs"].([]interface{}); !ok {
			t.Errorf("%s: PodIPs = %v, want a list", c.name, got["PodIPs"])
		}
	}
}
