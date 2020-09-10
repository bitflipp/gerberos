package main

type source interface {
	initialize(ps []string) error
	entries() (chan entry, error)
}

type fileSource struct {
	path string
}

type systemdSource struct {
	service string
}
