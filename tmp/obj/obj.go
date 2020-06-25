package main

import (
	"fmt"
)

type Human struct {
	name      string
	age       int
	nicknames []string
	dog       Dog
}

func (h *Human) setName(name string) {
	h.name = name
}

func (h *Human) getName() string {
	return h.name
}

func (h *Human) giveDog(d Dog) {
	h.dog = d
}

func (h *Human) getDog() Dog {
	return h.dog
}

func main() {
	h := &Human{}
	h.setName("Jason")
	fmt.Printf("Name: %s\n", h.getName())

	d := &Dog{}
	d.setName("foo")
	h.giveDog(*d)

	fmt.Printf("%v", h.getDog())
}
