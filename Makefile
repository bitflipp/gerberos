TAG := $(shell git describe --tags $(shell git rev-list --tags --max-count=1))

all: run

clean:
	rm -rf dist
	rm -rf *.coverage

dist: clean
	mkdir dist
	CGO_ENABLED=0 go build -o dist/gerberos -ldflags "-X main.tag=$(TAG)"

release: dist
	cp -r licenses-third-party gerberos.toml gerberos.service LICENSE dist
	cd dist && tar czvf gerberos-$(TAG).tar.gz *

run: dist
	dist/gerberos

test: clean
	go test -v -coverprofile=gerberos.coverage
	go tool cover -html=gerberos.coverage

test_system: clean
	go test -v -tags=system -coverprofile=gerberos.coverage
	go tool cover -html=gerberos.coverage

tag:
	@echo $(TAG)
