.PHONY: deps build package clean all

all: package

deps:
	go get -d -v -t ./...
	yarn

build: deps
	go build
	yarn start

package: build
	tar czvf gomibako.tar.gz ./gomibako templates static

clean:
	rm -rf gomibako.tar.gz gomibako static/script/ static/style/
