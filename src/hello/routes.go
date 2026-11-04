package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	config "github.com/jasonmkurtz/hellohttp/src/lib/config"
	db "github.com/jasonmkurtz/hellohttp/src/lib/db"
	jregex "github.com/jasonmkurtz/hellohttp/src/lib/jregex"
	routetypes "github.com/jasonmkurtz/hellohttp/src/lib/routes"
)

var app routetypes.Routes

// Defaults to 8080 rather than 5000, which macOS's AirPlay Receiver holds.
func getPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return port
	}

	return "8080"
}

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
	app = routetypes.Routes{Port: getPort(), Primary: HandleHello, Missing: Missing}
	app.Routes = []routetypes.Route{
		routetypes.Route{Route: "/hello", Handler: HandleHello},
		routetypes.Route{Route: "^/greet/?(?P<name>[a-zA-Z]+)?/?$", Handler: Greet, DenyPost: true},
		routetypes.Route{Route: "/read", Handler: Read},
		routetypes.Route{Route: "/newname", Handler: AddName, DenyGet: true},
		routetypes.Route{Route: "/config", Handler: ReadConfig},
		routetypes.Route{Route: "/whoami", Handler: WhoAmI},
		routetypes.Route{Route: "/foo", Handler: Bar},
		routetypes.Route{Route: "/version", Handler: Version},
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

// version is stamped in at build time (the git commit, from CI):
//
//	go build -ldflags "-X main.version=$(git rev-parse HEAD)" ./src/hello
var version = "dev"

func Version(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	fmt.Fprint(w, version)
}

func Bar(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	fmt.Fprintf(w, "This is /foo!")
}

// WhoAmI describes the pod and the connection it's answering, including
// which IP family (IPv4 or IPv6) the request came in on.
//
// ClientIP is the source address as the pod sees it, which isn't always the
// real client: with a NodePort, the node may rewrite it to its own address
// before forwarding. ServerIP is the pod address the connection arrived at.
func WhoAmI(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	host := os.Getenv("HOSTNAME")
	image := os.Getenv("MY_IMAGE")

	client := hostIP(r.RemoteAddr)
	var server net.IP
	if addr, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
		server = hostIP(addr.String())
	}

	w.Header().Set("Content-type", "application/json")
	ret := map[string]interface{}{
		"Host":         host,
		"Image":        image,
		"Version":      version,
		"ClientIP":     ipString(client),
		"ClientFamily": ipFamily(client),
		"ServerIP":     ipString(server),
		"ServerFamily": ipFamily(server),
		"PodIPs":       podIPs(),
	}
	js, err := json.Marshal(ret)
	if err != nil {
		fmt.Fprintf(w, "Error! %s", err)
		return
	}

	w.Write(js)
}

// hostIP returns the IP in a "host:port" address (or a bare IP), or nil if
// there isn't one.
func hostIP(addr string) net.IP {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	return net.ParseIP(host)
}

func ipString(ip net.IP) string {
	if ip == nil {
		return ""
	}

	return ip.String()
}

// ipFamily reports "IPv4" or "IPv6", or "" for no address. An IPv4-mapped
// IPv6 address (::ffff:192.0.2.1), which is how a dual-stack socket can see
// an IPv4 client, counts as IPv4.
func ipFamily(ip net.IP) string {
	switch {
	case ip == nil:
		return ""
	case ip.To4() != nil:
		return "IPv4"
	default:
		return "IPv6"
	}
}

// podIPs lists the addresses on this pod's network interfaces, leaving out
// loopback and link-local (169.254.x.x, fe80::) ones, which every pod has.
// On a dual-stack cluster that's one IPv4 and one IPv6 address.
func podIPs() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	ips := []string{}
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() || ipnet.IP.IsLinkLocalUnicast() {
			continue
		}
		ips = append(ips, ipnet.IP.String())
	}

	return ips
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

// findGreeting looks up a custom greeting for name, falling back to "Hey"
// if there isn't one or the database is unreachable.
func findGreeting(name string) string {
	db := app.Database.Open()
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var greeting sql.NullString
	err := db.QueryRowContext(ctx, "SELECT greeting FROM names WHERE name=?", name).Scan(&greeting)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("greeting lookup for %s failed: %v\n", name, err)
		return "Hey"
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
	reply, err := backend.Request("api/foo")
	if err != nil {
		fmt.Printf("backend request failed: %v\n", err)
		return
	}
	fmt.Fprintf(w, "Backend says: \"%s\"", reply)
}

func Missing(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	w.WriteHeader(404)
	fmt.Fprintf(w, "%s was requested but not found.", r.URL.Path)
}

func HandleHello(w http.ResponseWriter, r *http.Request, route routetypes.Route) {
	fmt.Fprintf(w, "Hello, this might be /hello!")
}
