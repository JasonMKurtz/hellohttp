package main

import (
	"fmt"
	"net/http"
)

func main() {
	r := &Routes{port: "80"}
	r.routes = []Route{
		Route{"/api/foo", HandleFoo},
	}

	r.Listen()
}

func HandleFoo(w http.ResponseWriter, r *http.Request, route string) {
	fmt.Fprintf(w, "You're at %s on the backend.", route)
}
