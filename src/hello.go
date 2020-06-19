package main

import (
    "fmt"
    "net/http"
    "strings"
    "os"
)

func main() {
    http.HandleFunc("/", HelloServer)
    http.ListenAndServe(":8080", nil)
}

func GetIP(r *http.Request) string {
    forwarded := r.Header.Get("X-FORWARDED-FOR")
    if forwarded != "" {
        return forwarded
    }
    return strings.Split(r.RemoteAddr, ":")[0]
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
    // non-tagged commit test
    host, _ := os.Hostname()
    fmt.Fprintf(w, "Hello, this is %s, and you are %s\n", host, GetIP(r))
    fmt.Fprintf(w, "This is v12.")
}
