package main

import (
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
