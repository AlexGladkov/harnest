VERSION := 0.1.0
BINARY := harnest
MODULE := github.com/AlexGladkov/harnest

.PHONY: build clean release

build:
	go build -o $(BINARY) ./cmd/harnest/

clean:
	rm -rf $(BINARY) dist/

release: clean
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY)-darwin-amd64 ./cmd/harnest/
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY)-darwin-arm64 ./cmd/harnest/
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY)-linux-amd64 ./cmd/harnest/
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY)-linux-arm64 ./cmd/harnest/
	cd dist && shasum -a 256 * > checksums.txt
	@echo "Release binaries in dist/"
