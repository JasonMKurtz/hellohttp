package main

import (
	"fmt"
	"net/http"
)

func main() {
	r := &Routes{port: "9001"}
	r.routes = []Route{
		Route{"/api/foo", HandleFoo},
	}

	r.Listen()
}

func HandleFoo(w http.ResponseWriter, r *http.Request, route string) {
	fmt.Fprintf(w, "You've reached /foo on the backend.")
}
