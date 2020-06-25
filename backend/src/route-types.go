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

func (r *Routes) Listen() {
	fmt.Printf("Listening on :%s...\n", r.port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", r.port), r)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
