package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	utils "../utils"
)

type Config struct {
	Filename string
	Valid    bool
	Content  string
}

func Get(file string) *Config {
	c := &Config{Filename: file, Content: "", Valid: false}
	filename := fmt.Sprintf("/config/%s", file)
	f, err := os.Stat(filename)
	if err != nil || f.IsDir() {
		return c
	}

	value, err := ioutil.ReadFile(filename)
	if err != nil {
		return c
	}

	c.Content = string(value)
	c.Valid = true

	return c
}

func IsRouteDenied(route string) bool {
	config := Get("denyroutes")
	fmt.Printf("Config: %v\n", config.Content)
	if !config.Valid {
		return false
	}

	routes := strings.Split(config.Content, " ")
	fmt.Printf("Checking against %s\n", route)
	if utils.InList(routes, route) {
		fmt.Printf("Route %s in deny list.\n", route)
		return true
	}

	fmt.Printf("Route %s not in deny list.\n", route)
	return false
}
