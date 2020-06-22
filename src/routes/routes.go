package main

import (
	"fmt"
	"net/http"
)

type Route struct {
	route   string
	handler func(w http.ResponseWriter, r *http.Request)
}

type Routes struct {
	routes []Route
	port   string
}

func (r *Routes) addRoute(route Route) {
	r.routes = append(r.routes, route)
}

func (r *Routes) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		if req.URL.Path == route.route {
			route.handler(w, req)
			return
		}
	}

	http.NotFound(w, req)
	return
}

func main() {
	r := &Routes{port: "8000"}
	r.addRoute(Route{route: "/hello", handler: HandleHello})
	r.addRoute(Route{route: "/bar", handler: HandleBar})

	fmt.Printf("Listening on :%s...\n", r.port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", r.port), r)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func HandleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, this is /hello!")
}

func HandleBar(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, this is /bar!")
}
