package main

import (
	"fmt"
	"net/http"
	"regexp"
)

type RouteHandler func(w http.ResponseWriter, r *http.Request, route string)

type Route struct {
	route   string
	handler RouteHandler
}

type Routes struct {
	routes  []Route
	port    string
	primary RouteHandler
	missing RouteHandler
}

func (r *Routes) addRoute(route Route) {
	r.routes = append(r.routes, route)
}

func (r *Routes) addDefaultRoute(h RouteHandler) {
	r.primary = h
}

func (r *Routes) add404Route(h RouteHandler) {
	r.missing = h
}

func (r *Routes) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		reg, _ := regexp.Compile(route.route)
		path := req.URL.Path
		if path == "/" && r.primary != nil {
			r.primary(w, req, "/")
			return
		} else if reg.MatchString(path) {
			route.handler(w, req, route.route)
			return
		}
	}

	if r.missing != nil {
		r.missing(w, req, "")
		return
	}

	http.NotFound(w, req)
	return
}

func main() {
	r := &Routes{port: "8080", primary: HandleHello, missing: Missing}
	r.addRoute(Route{route: "/hello", handler: HandleHello})
	r.addRoute(Route{route: "/bar", handler: HandleBar})
	r.addRoute(Route{route: "^/greet/(?P<name>.+)$", handler: Greet})

	fmt.Printf("Listening on :%s...\n", r.port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", r.port), r)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func Greet(w http.ResponseWriter, r *http.Request, route string) {
	reg := &JRegex{route, r.URL.Path}
	fmt.Fprintf(w, "Hello %s!\n", reg.GetNamedGroups()["name"])
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
