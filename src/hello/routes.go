package main

import (
	"fmt"
	"net/http"

	db "../lib/db"
	jregex "../lib/jregex"
	routetypes "../lib/routes"
)

var app routetypes.Routes

func main() {
	app = routetypes.Routes{Port: "8080", Primary: HandleHello, Missing: Missing}
	app.Routes = []routetypes.Route{
		routetypes.Route{Route: "/hello", Handler: HandleHello},
		routetypes.Route{Route: "/bar", Handler: HandleBar},
		routetypes.Route{Route: "^/greet/(?P<name>[a-zA-Z]+)$", Handler: Greet},
		routetypes.Route{Route: "/read", Handler: Read},
	}

	app.AddService("hellohttp-backend", 80)

	app.Listen()
}

func Read(w http.ResponseWriter, r *http.Request, route string) {
	d := db.Database{
		Host: "mysql-1593208582",
		Port: "3306",
		User: "root",
		Db:   "",
	}

	res := d.Read()

	fmt.Fprintf(w, "%v", res)
}

func Greet(w http.ResponseWriter, r *http.Request, route string) {
	reg := &jregex.JRegex{Exp: route, Haystack: r.URL.Path}
	fmt.Fprintf(w, "Hello %s!\n", reg.GetNamedGroups()["name"])

	backend := app.Services["hellohttp-backend"]
	fmt.Fprintf(w, "Backend says: \"%s\"", backend.Request("api/foo"))
}

func Missing(w http.ResponseWriter, r *http.Request, route string) {
	fmt.Fprintf(w, r.URL.Path+" was requested but not found.")
}

func HandleHello(w http.ResponseWriter, r *http.Request, route string) {
	fmt.Fprintf(w, "Hello, this might be /hello!")
}

func HandleBar(w http.ResponseWriter, r *http.Request, route string) {
	fmt.Fprintf(w, "Hello, this is /bar!")
}
