all: build

deps:
	go get -u github.com/benbjohnson/ego/cmd/ego
	go get -u github.com/shurcooL/vfsgen/cmd/vfsgendev

clean:
	rm -f data/assets_vfsdata.go
	rm -f data/template/*.ego.go
	rm -f ulvemelk

assets:
	cd data && go generate

templates:
	ego data/template

build: clean assets templates
	go build

dev: templates
	go run -tags dev cmd/ulvemelk/ulvemelk.go