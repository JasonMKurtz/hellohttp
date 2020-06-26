package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

type RouteHandler func(w http.ResponseWriter, r *http.Request, route string)

type Route struct {
	route   string
	handler RouteHandler
}

type Routes struct {
	routes   []Route
	port     string
	primary  RouteHandler
	missing  RouteHandler
	services map[string]Service
}

type Service struct {
	name string
	port int
}

func (s *Service) getString() string {
	return fmt.Sprintf("http://%s:%d", s.name, s.port)
}

func (s *Service) request(path string) string {
	resp, err := http.Post(fmt.Sprintf("%s/%s", s.getString(), path))
	if err != nil {
		fmt.Printf("%v\n", err)
		panic(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
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

func (r *Routes) addService(name string, port int) {
	if r.services == nil {
		r.services = make(map[string]Service)
	}

	r.services[name] = Service{name, port}
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
