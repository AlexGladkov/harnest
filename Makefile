VERSION := 0.12.0
BINARY := harnest
MODULE := github.com/AlexGladkov/harnest

.PHONY: build clean release

build:
	go build -o $(BINARY) ./cmd/harnest/

clean:
	rm -f $(BINARY) $(BINARY).exe
	rm -rf dist/

release: clean
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY)-darwin-amd64 ./cmd/harnest/
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY)-darwin-arm64 ./cmd/harnest/
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY)-linux-amd64 ./cmd/harnest/
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY)-linux-arm64 ./cmd/harnest/
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY)-windows-amd64.exe ./cmd/harnest/
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY)-windows-arm64.exe ./cmd/harnest/
	cd dist && (shasum -a 256 * > checksums.txt 2>/dev/null || sha256sum * > checksums.txt)
	@echo "Release binaries in dist/"
