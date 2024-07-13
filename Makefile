dist: # output directory for artifacts
	mkdir -p dist
	rm -rf dist/*

build: dist # Build binaries
	go build -o dist/logues cmd/server/server.go

exec: build # build & run server
	dist/logues

image: # build container image
	podman build -f container/Containerfile -t logues:0.0.1 .

container: image # build & run container image
	podman run --name logues-local -p 8080:7331 -d logues:0.0.1

test: # test all
	go test ./... -v 

race: # test all with race conditions
	go test ./... -v -race

coverage: # test code coverage 
	go test -coverprofile=coverage.out ./...
	go tool cover -html coverage.out

.PHONY: dist container
