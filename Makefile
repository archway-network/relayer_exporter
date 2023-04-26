BUILD_FOLDER = dist

.PHONY: all
all: install

.PHONY: install
install:
	go install cmd/relayer_exporter/relayer_exporter.go

.PHONY: build
build:
	@goreleaser build --single-target --config .goreleaser.yaml --snapshot --clean

.PHONY: clean
clean:
	rm -rf $(BUILD_FOLDER)

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
