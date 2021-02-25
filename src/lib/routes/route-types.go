package routetypes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	config "github.com/jasonmkurtz/hellohttp/src/lib/config"
	db "github.com/jasonmkurtz/hellohttp/src/lib/db"
	jregex "github.com/jasonmkurtz/hellohttp/src/lib/jregex"
	utils "github.com/jasonmkurtz/hellohttp/src/lib/utils"
)

type RouteHandler func(w http.ResponseWriter, r *http.Request, route Route)

type Route struct {
	Route    string
	Handler  RouteHandler
	DenyGet  bool
	DenyPost bool
}

type Routes struct {
	Routes     []Route
	Port       string
	Primary    RouteHandler
	Missing    RouteHandler
	Services   map[string]Service
	Database   db.Database
	ConfigPath string
}

type Service struct {
	name string
	port int
}

func (s *Service) getString() string {
	return fmt.Sprintf("http://%s:%d", s.name, s.port)
}

func (s *Service) Request(path string) string {
	resp, err := http.Get(fmt.Sprintf("%s/%s", s.getString(), path))
	if err != nil {
		fmt.Printf("%v\n", err)
		panic(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func (r *Routes) SetConfigPath(p string) {
	r.ConfigPath = p
}

func (r *Routes) AddDefaultRoute(h RouteHandler) {
	r.Primary = h
}

func (r *Routes) Add404Route(h RouteHandler) {
	r.Missing = h
}

func (r *Routes) AddService(name string, port int) {
	if r.Services == nil {
		r.Services = make(map[string]Service)
	}

	r.Services[name] = Service{name, port}
}

func (r *Routes) AddDatabase(d db.Database) {
	r.Database = d
}

func (r *Routes) missing(w http.ResponseWriter, req *http.Request) {
	if r.Missing == nil {
		http.NotFound(w, req)
		return
	}

	r.Missing(w, req, Route{})
}

func (r *Routes) allowMethod(w http.ResponseWriter, req *http.Request, route Route) bool {
	noget := route.DenyGet
	nopost := route.DenyPost

	if noget && req.Method == "GET" {
		return false
	}

	if nopost && req.Method == "POST" {
		return false
	}

	return true
}

func (r *Route) IsRouteDenied(route, configpath string) bool {
	config := config.Get(fmt.Sprintf("%s/denyroutes", configpath))
	if !config.Valid {
		return false
	}

	routes := strings.Split(config.Content, " ")
	if utils.InList(routes, route) || jregex.IsMatch(route, r.Route) {
		return true
	}

	return false
}

func (r *Routes) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.Routes {
		reg, _ := regexp.Compile(route.Route)
		path := req.URL.Path
		if path == "/" && r.Primary != nil {
			r.Primary(w, req, route)
			return
		} else if reg.MatchString(path) {
			/*
				if !r.allowMethod(w, req, route) || route.IsRouteDenied(path, r.ConfigPath) {
					r.missing(w, req)
					return
				}
			*/

			route.Handler(w, req, route)
			return
		}
	}

	r.missing(w, req)
	return
}

func (r *Routes) Listen() {
	fmt.Printf("Listening on :%s...\n", r.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", r.Port), r)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
