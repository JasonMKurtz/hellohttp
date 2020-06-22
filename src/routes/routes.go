package main

import (
	"fmt"
	"net/http"
)

type RouteHandler func(w http.ResponseWriter, r *http.Request)

type Route struct {
	route   string
	handler RouteHandler
}

type Routes struct {
	routes  []Route
	port    string
	missing RouteHandler
}

func (r *Routes) addRoute(route Route) {
	r.routes = append(r.routes, route)
}

func (r *Routes) addDefaultRoute(h RouteHandler) {
	r.addRoute(Route{
		route:   "/",
		handler: h,
	})
}

func (r *Routes) add404Route(h RouteHandler) {
	r.missing = h
}

func (r *Routes) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		if req.URL.Path == route.route {
			route.handler(w, req)
			return
		}
	}

	if r.missing != nil {
		r.missing(w, req)
		return
	}

	http.NotFound(w, req)
	return
}

func main() {
	r := &Routes{port: "8080"}
	r.addDefaultRoute(HandleHello)
	r.add404Route(Missing)
	r.addRoute(Route{route: "/hello", handler: HandleHello})
	r.addRoute(Route{route: "/bar", handler: HandleBar})

	fmt.Printf("Listening on :%s...\n", r.port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", r.port), r)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func Missing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, r.URL.Path+" was requested but not found.")
}

func HandleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, this might be /hello!")
}

func HandleBar(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, this is /bar!")
}
