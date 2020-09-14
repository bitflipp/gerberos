all: run

clean:
	rm -rf dist
	rm -rf *.coverage

build: clean
	mkdir dist
	CGO_ENABLED=0 go build -o dist/gerberos

run: build
	sudo dist/gerberos

test: clean
	go test -coverprofile=gerberos.coverage
	go tool cover -html=gerberos.coverage