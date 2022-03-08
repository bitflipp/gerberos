all: run

clean:
	rm -rf dist
	rm -rf *.coverage

dist: clean
	mkdir dist
	CGO_ENABLED=0 go build -o dist/gerberos

release: dist
	cp -r licenses-third-party gerberos.toml gerberos.service LICENSE dist
	cd dist && tar czvf gerberos-$(shell cat VERSION).tar.gz *

run: dist
	dist/gerberos

test: clean
	go test -v -tags=system -coverprofile=gerberos.coverage
	go tool cover -html=gerberos.coverage
