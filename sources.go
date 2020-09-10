package main

type source interface {
	initialize(ps []string) error
	matches() (chan match, error)
}

type fileSource struct {
	path string
}

type systemdSource struct {
	service string
}
