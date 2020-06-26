package main

import (
	"fmt"
	"net/http"
)

var app Routes

func main() {
	app = Routes{port: "8080", primary: HandleHello, missing: Missing}
	app.routes = []Route{
		Route{"/hello", HandleHello},
		Route{"/bar", HandleBar},
		Route{"^/greet/(?P<name>[a-zA-Z]+)$", Greet},
	}

	app.addService("hellohttp-backend", 80)

	app.Listen()
}

func Greet(w http.ResponseWriter, r *http.Request, route string) {
	reg := &JRegex{route, r.URL.Path}
	fmt.Fprintf(w, "Hello %s!\n", reg.GetNamedGroups()["name"])

	backend := app.services["hellohttp-backend"]
	fmt.Fprintf(w, "Backend says: \"%s\"", backend.request("api/foo"))
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
