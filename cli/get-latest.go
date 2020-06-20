package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
	//"math"
    "net/http"
	//"strings"
	//"strconv"
)

type DockerHubToken struct {
    Token string
    Expires_In int
    Issued_At string
}

type DockerHubRepoInfo struct {
    Name string
    Tags []string
}

func notmain() {
    resp := `{"token": "abcd", "expires_in": 300, "issued_at": "now"}`
    var token DockerHubToken
    json.Unmarshal([]byte(resp), &token)
    fmt.Printf("Token: %s, expires: %d, issued: %s\n", token.Token, token.Expires_In, token.Issued_At)
}

    
func Filter(vs []string, f func(string) bool) []string {
    vsf := make([]string, 0)
    for _, v := range vs {
        if f(v) {
            vsf = append(vsf, v)
        }
    }
    return vsf
}

func Max(tags []string) string {
	maxver := "0.0"
	for _, v := range tags {
		//tag, _ := strconv.ParseFloat(v, 64)
		//tag = tag * 100
		//fmt.Printf("%d %d\n", tag, maxver)
		if maxver < v {
			maxver = v
			//fmt.Println(maxver)
		}
	}

	fmt.Println(maxver)
	return maxver
}

func main() {
    //var token string = GetToken().Token
    var info DockerHubRepoInfo
    resp := `{"name":"jmliber/hellohttp","tags":["0.1","0.11","0.12","0.2","0.3","0.4","latest"]}`
    json.Unmarshal([]byte(resp), &info)
	ignoreLatest := Filter(info.Tags, func(v string) bool { return v != "latest" })
	latest := Max(ignoreLatest)
	fmt.Println(latest)
}

func GetToken() DockerHubToken {
    resp, err := http.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:jmliber/hellohttp:pull")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    text, err := ioutil.ReadAll(resp.Body)

    var token DockerHubToken
    json.Unmarshal(text, &token)
    return token
}
