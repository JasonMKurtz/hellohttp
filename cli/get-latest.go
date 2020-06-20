package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DockerHubToken struct {
	Token      string
	Expires_In int
	Issued_At  string
}

type DockerHubRepoInfo struct {
	Name string
	Tags []string
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
		if v > maxver {
			maxver = v
		}
	}

	return maxver
}

func GetTags(token string) []string {
	req, _ := http.NewRequest("GET", "https://registry.hub.docker.com/v2/jmliber/hellohttp/tags/list", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	text, _ := ioutil.ReadAll(resp.Body)

	var taginfo DockerHubRepoInfo
	json.Unmarshal(text, &taginfo)

	return taginfo.Tags
}

func main() {
	tags := GetTags(GetToken().Token)
	ignoreLatest := Filter(tags, func(v string) bool { return v != "latest" })
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
