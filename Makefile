all: run

clean:
	rm -rf dist

build: clean
	mkdir dist
	go build -o dist/gerberos

run: build
	sudo dist/gerberos
