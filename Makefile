BUILD_FOLDER = build
DIST_FOLDER = dist
GORELEASER_VERSION = v1.21.0
DOCKER := $(shell which docker)
PACKAGE_NAME = github.com/archway-network/relayer_exporter

.PHONY: all
all: install

.PHONY: install
install: go.sum
	go install cmd/relayer_exporter/relayer_exporter.go

.PHONY: go.sum
go.sum:
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

.PHONY: build
build: go.sum
	go build -o build/relayer_exporter ./cmd/relayer_exporter/relayer_exporter.go

.PHONY: clean
clean:
	rm -rf $(BUILD_FOLDER)
	rm -rf $(DIST_FOLDER)

.PHONY: test
test:
	go test -race ./...

.PHONY: text-cover
test-cover:
	go test -race ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

.PHONY: test-ci
test-ci:
	@go get ./...
	@go test ./...

.PHONY: lint
lint:
	@golangci-lint run
	@go mod verify

.PHONY: update
update:
	go get -u -d ./...
	@go mod tidy
	@go build -o "$(TMPDIR)/relayer_exporter" cmd/relayer_exporter/relayer_exporter.go
	@git diff -- go.mod go.sum

release-dryrun:
	$(DOCKER) run \
		--rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		goreleaser/goreleaser-cross:$(GORELEASER_VERSION) \
		--skip-publish \
		--clean \
		--skip-validate

release:
	$(DOCKER) run \
		--rm \
		-e GITHUB_TOKEN="$(GITHUB_TOKEN)" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		goreleaser/goreleaser-cross:$(GORELEASER_VERSION) \
		--clean
