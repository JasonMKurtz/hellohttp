package main

type Dog struct {
	name string
}

func (d *Dog) setName(name string) {
	d.name = name
}

func (d *Dog) getName() string {
	return d.name
}
