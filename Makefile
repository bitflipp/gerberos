all: run

clean:
	rm -rf dist
	rm -rf *.coverage

dist: clean
	mkdir dist
	CGO_ENABLED=0 go build -o dist/gerberos

package: dist
	cp gerberos.toml gerberos.service dist
	cd dist && tar czvf gerberos.tar.gz *

run: build
	sudo dist/gerberos

test: clean
	go test -coverprofile=gerberos.coverage
	go tool cover -html=gerberos.coverage
