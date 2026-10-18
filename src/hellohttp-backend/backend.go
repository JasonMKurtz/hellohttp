package main

import (
	"fmt"
	"net/http"
	"os"

	routetypes "github.com/jasonmkurtz/hellohttp/src/lib/routes"
)

// Same as hello: $PORT, or 8080.
func getPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return port
	}

	return "8080"
}

func main() {
	r := &routetypes.Routes{Port: getPort()}
	r.Routes = []routetypes.Route{
		routetypes.Route{Route: "/api/foo", Handler: HandleFoo},
	}

	r.Listen()
}

func HandleFoo(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	fmt.Fprintf(w, "You're at %s on the backend.", route.Route)
}
