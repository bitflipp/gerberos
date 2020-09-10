package main

type source interface {
	initialize([]string) error
	entries() (chan *entry, error)
}

type fileSource struct {
	path string
}

type systemdSource struct {
	service string
}
