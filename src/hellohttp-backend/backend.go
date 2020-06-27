package main

import (
	"fmt"
	"net/http"

	routetypes "../lib/routes"
)

func main() {
	r := &routetypes.Routes{Port: "80"}
	r.Routes = []routetypes.Route{
		routetypes.Route{Route: "/api/foo", Handler: HandleFoo},
	}

	r.Listen()
}

func HandleFoo(w http.ResponseWriter, r *http.Request, route string) {
	fmt.Fprintf(w, "You're at %s on the backend.", route)
}
