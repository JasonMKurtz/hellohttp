package routetypes

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

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

var serviceClient = &http.Client{Timeout: 2 * time.Second}

func (s *Service) Request(path string) (string, error) {
	resp, err := serviceClient.Get(fmt.Sprintf("%s/%s", s.getString(), path))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
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

// How long to keep serving after SIGTERM. Kubernetes sends SIGTERM as soon
// as a pod starts terminating, but the load balancer takes a few seconds to
// stop routing to it; exiting right away drops those requests.
var drainDelay = 15 * time.Second

func (r *Routes) Listen() {
	srv := &http.Server{Addr: fmt.Sprintf(":%s", r.Port), Handler: r}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)

	go func() {
		<-stop
		fmt.Printf("Shutting down in %s...\n", drainDelay)
		time.Sleep(drainDelay)

		// Stop accepting connections and let in-flight requests finish.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	fmt.Printf("Listening on :%s...\n", r.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("%s\n", err)
	}
}
