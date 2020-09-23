all: run

clean:
	rm -rf dist
	rm -rf *.coverage

dist: clean
	mkdir dist
	CGO_ENABLED=0 go build -o dist/gerberos

release: dist
	cp gerberos.toml gerberos.service dist
	wget -O dist/LICENCE_BurntSushi_toml https://raw.githubusercontent.com/BurntSushi/toml/master/COPYING
	wget -O dist/LICENCE_golang_go https://raw.githubusercontent.com/golang/go/master/LICENSE
	cd dist && tar czvf gerberos.tar.gz *

run: dist
	sudo dist/gerberos

test: clean
	go test -coverprofile=gerberos.coverage
	go tool cover -html=gerberos.coverage
