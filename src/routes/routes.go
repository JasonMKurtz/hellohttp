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
}

func (r *Routes) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range getRoutes().routes {
		if req.URL.Path == route.route {
			route.handler(w, req)
			return
		}
	}

	http.NotFound(w, req)
	return
}

func getRoutes() *Routes {
	routes := []Route{
		Route{route: "/hello", handler: HandleHello},
		Route{route: "/bar", handler: HandleBar},
	}

	return &Routes{routes: routes}
}

func main() {
	fmt.Printf("Listening...")
	http.ListenAndServe(":9000", getRoutes())
}

func HandleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, this is /hello!")
}

func HandleBar(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, this is /bar!")
}
