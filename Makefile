VERSION ?= dev
DIST_DIR ?= dist
PACKAGE = webguard-server-agent

.PHONY: build build-linux-amd64 build-linux-arm64 check fmt-check test vet dist deb clean

build:
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o bin/$(PACKAGE) ./cmd/$(PACKAGE)

build-linux-amd64:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o $(DIST_DIR)/$(PACKAGE)_$(VERSION)_linux_amd64 ./cmd/$(PACKAGE)

build-linux-arm64:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o $(DIST_DIR)/$(PACKAGE)_$(VERSION)_linux_arm64 ./cmd/$(PACKAGE)

fmt-check:
	test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))"

test:
	go test ./...

vet:
	go vet ./...

check: fmt-check vet test

dist: build-linux-amd64 build-linux-arm64
	cd $(DIST_DIR) && sha256sum $(PACKAGE)_$(VERSION)_linux_amd64 $(PACKAGE)_$(VERSION)_linux_arm64 > SHA256SUMS

deb:
	./scripts/build-deb.sh $(VERSION) amd64 $(DIST_DIR)/deb
	./scripts/build-deb.sh $(VERSION) arm64 $(DIST_DIR)/deb

clean:
	rm -rf bin $(DIST_DIR)
