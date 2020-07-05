package config

import (
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	Filename string
	Valid    bool
	Content  string
}

func Get(filename string) *Config {
	c := &Config{Filename: filename, Content: "", Valid: false}
	f, err := os.Stat(filename)
	if err != nil || f.IsDir() {
		fmt.Printf("Config: %s\nError? %v\nDir? %v", filename, err, f.IsDir())
		return c
	}

	value, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Can't read %s", filename)
		return c
	}

	c.Content = string(value)
	c.Valid = true

	return c
}
