package main

import (
	"fmt"
	"net/http"

	routetypes "../lib/routes"
)

func main() {
	app := routetypes.Routes{Port: "80", Primary: Load}
	app.Routes = []routetypes.Route{
		routetypes.Route{Route: "/load", Handler: Load},
	}
	app.Listen()
}

func Load(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	var sum int
	for i := 0; i <= 1000000000000000; i++ {
		sum = sum + (i % 2) - 3
	}
	fmt.Fprintf(w, "Okay!")
}
