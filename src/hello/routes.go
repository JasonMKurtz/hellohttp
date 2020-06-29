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
		routetypes.Route{Route: "^/greet/(?P<name>[a-zA-Z]+)$", Handler: Greet, DenyPost: true},
		routetypes.Route{Route: "/read", Handler: Read},
		routetypes.Route{Route: "/post", Handler: TestPost, DenyGet: true},
	}

	app.AddService("hellohttp-backend", 80)

	app.AddDatabase(db.Database{
		Host: "mysql",
		Port: "3306",
		User: "root",
		Db:   "hello",
	})

	app.Listen()
}

func TestPost(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	fmt.Fprintf(w, "POST!")
}

type Name struct {
	name string
	foo  string
}

func Read(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	q := app.Database.Query("SELECT name FROM hello")
	var names []Name
	for q.Next() {
		var n Name
		err := q.Scan(&n.name)
		if err != nil {
			panic(err)
		}

		names = append(names, n)
	}

	for _, v := range names {
		fmt.Fprintf(w, "Name: %s\n", v.name)
	}
}

func Greet(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	reg := &jregex.JRegex{Exp: route.Route, Haystack: r.URL.Path}
	fmt.Fprintf(w, "Hello %s!\n", reg.GetNamedGroups()["name"])

	backend := app.Services["hellohttp-backend"]
	fmt.Fprintf(w, "Backend says: \"%s\"", backend.Request("api/foo"))
}

func Missing(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	fmt.Fprintf(w, r.URL.Path+" was requested but not found.")
}

func HandleHello(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	fmt.Fprintf(w, "Hello, this might be /hello!")
}

func HandleBar(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	fmt.Fprintf(w, "Hello, this is /bar!")
}
