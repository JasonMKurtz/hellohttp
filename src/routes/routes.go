package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

var app Routes

func main() {
	app = Routes{port: "8080", primary: HandleHello, missing: Missing}
	app.routes = []Route{
		Route{"/hello", HandleHello},
		Route{"/bar", HandleBar},
		Route{"^/greet/(?P<name>.+)$", Greet},
	}

	app.addService("hellohttp-backend", 9001)

	app.Listen()
}

func Greet(w http.ResponseWriter, r *http.Request, route string) {
	reg := &JRegex{route, r.URL.Path}
	fmt.Fprintf(w, "Hello %s!\n", reg.GetNamedGroups()["name"])

	backend := app.services["hellohttp-backend"]

	resp, err := http.Get(fmt.Sprintf("%s/api/foo", backend.getString()))
	if err != nil {
		fmt.Printf("%v\n", err)
		panic(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Fprintf(w, "Backend response: %s\n", body)
}

func Missing(w http.ResponseWriter, r *http.Request, route string) {
	fmt.Fprintf(w, r.URL.Path+" was requested but not found.")
}

func HandleHello(w http.ResponseWriter, r *http.Request, route string) {
	fmt.Fprintf(w, "Hello, this might be /hello!")
}

func HandleBar(w http.ResponseWriter, r *http.Request, route string) {
	fmt.Fprintf(w, "Hello, this is /bar!")
}
