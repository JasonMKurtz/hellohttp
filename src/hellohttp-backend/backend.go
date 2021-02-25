package main

import (
	"fmt"
	"net/http"

	routetypes "github.com/jasonmkurtz/hellohttp/src/lib/routes"
)

func main() {
	r := &routetypes.Routes{Port: "80"}
	r.Routes = []routetypes.Route{
		routetypes.Route{Route: "/api/foo", Handler: HandleFoo},
	}

	r.Listen()
}

func HandleFoo(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	fmt.Fprintf(w, "You're at %s on the backend.", route.Route)
}
