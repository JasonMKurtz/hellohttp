package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	config "../lib/config"
	db "../lib/db"
	jregex "../lib/jregex"
	routetypes "../lib/routes"
)

var app routetypes.Routes

func getMysqlHost() string {
	host := os.Getenv("MYSQL_HOST")
	if host != "" {
		return host
	}

	return "mysql"
}

func getConfigPath() string {
	path := os.Getenv("CONFIG_PATH")
	if path != "" {
		return path
	}

	return "/config"
}

func main() {
	app = routetypes.Routes{Port: "8080", Primary: HandleHello, Missing: Missing}
	app.Routes = []routetypes.Route{
		routetypes.Route{Route: "/hello", Handler: HandleHello},
		routetypes.Route{Route: "^/greet/?(?P<name>[a-zA-Z]+)?/?$", Handler: Greet, DenyPost: true},
		routetypes.Route{Route: "/read", Handler: Read},
		routetypes.Route{Route: "/newname", Handler: AddName, DenyGet: true},
		routetypes.Route{Route: "/config", Handler: ReadConfig},
		routetypes.Route{Route: "/whoami", Handler: WhoAmI},
	}

	app.AddService("hellohttp-backend", 80)

	app.AddDatabase(db.Database{
		Host: getMysqlHost(),
		Port: "3306",
		User: "root",
		Db:   "hello",
	})

	app.SetConfigPath(getConfigPath())

	fmt.Printf("Using database %s\nConfig Path: %s\n", getMysqlHost(), getConfigPath())

	app.Listen()
}

func WhoAmI(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	host := os.Getenv("HOSTNAME")
	w.Header().Set("Content-type", "application/json")
	ret := map[string]string{
		"Host": host,
	}
	fmt.Fprintf(w, "%v", ret)
}

type Name struct {
	name string
	foo  string
}

func ReadConfig(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	config := config.Get("nosuch")
	data := config.Content
	if !config.Valid {
		data = "invalid"
	}

	fmt.Fprintf(w, "Data: %s.\n", data)
}

func findGreeting(name string) string {
	db := app.Database.Open()

	var greeting sql.NullString
	err := db.QueryRow("SELECT greeting FROM names WHERE name=?", name).Scan(&greeting)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if greeting.Valid && greeting.String != "" {
		return greeting.String
	} else {
		return "Hey"
	}
}

func Read(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	q := app.Database.Query("SELECT name FROM names")
	var names []Name
	for q.Next() {
		var n Name
		err := q.Scan(&n.name)
		if err != nil {
			panic(err)
		}

		names = append(names, n)
	}

	fmt.Fprintf(w, "So far, we've got...\n")
	for _, v := range names {
		fmt.Fprintf(w, "Name: %s\n", v.name)
	}
}

func AddName(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	db := app.Database.Open()
	q, err := db.Prepare("INSERT INTO names (`name`, `greeting`) values(?, ?)")
	if err != nil {
		panic(err)
	}

	name := r.URL.Query().Get("name")
	greeting := r.URL.Query().Get("greeting")
	if name == "" {
		panic(err)
	}

	q.Exec(name, greeting)

	fmt.Fprintf(w, "%s added!\n", name)
}

func Greet(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	reg := &jregex.JRegex{Exp: route.Route, Haystack: r.URL.Path}
	name := reg.GetNamedGroups()["name"]
	if name == "" {
		fmt.Fprintf(w, "Hey guest!\n")
	} else {
		fmt.Fprintf(w, "%s %s!\n", findGreeting(name), name)
	}

	backend := app.Services["hellohttp-backend"]
	fmt.Fprintf(w, "Backend says: \"%s\"", backend.Request("api/foo"))
}

func Missing(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	fmt.Fprintf(w, "%s was requested but not found.", r.URL.Path)
}

func HandleHello(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	fmt.Fprintf(w, "Hello, this might be /hello!")
}
