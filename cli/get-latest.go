package main

import (
	"encoding/json"
    "flag"
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
    Errors []string
}

func Filter(vs []string, f func(string) bool) []string {
    var vsf []string
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func Max(tags []string) string {
    if len(tags) == 0 {
        return ""
    }

	maxver := "-1"
	for _, v := range tags {
		if v > maxver {
			maxver = v
		}
	}

	return maxver
}

func GetTags(image string, token string) []string {
    url := fmt.Sprintf("https://registry.hub.docker.com/v2/jmliber/%s/tags/list", image)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client := &http.Client{}
	resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }

	defer resp.Body.Close()

	text, _ := ioutil.ReadAll(resp.Body)

	var taginfo DockerHubRepoInfo
	json.Unmarshal(text, &taginfo)

	return taginfo.Tags
}

func main() {
    image := flag.String("image", "hellohttp", "The image to pull tags from.")
    flag.Parse()

    token := GetToken(*image).Token
	tags := GetTags(*image, token)

	ignoreLatest := Filter(tags, func(v string) bool { return v != "latest" })
	latest := Max(ignoreLatest)
	fmt.Println(latest)
}

func GetToken(image string) DockerHubToken {
    url := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:jmliber/%s:pull", image)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	text, err := ioutil.ReadAll(resp.Body)

	var token DockerHubToken
	json.Unmarshal(text, &token)
	return token
}
