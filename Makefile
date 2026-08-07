GOCMD?= go
GOOS=$(shell $(GOCMD) env GOOS)
GOARCH=$(shell $(GOCMD) env GOARCH)
GOTESTARCH?=$(GOARCH)

GO_TOOL  = GOOS= GOARCH= $(GOCMD) tool
MDATAGEN = $(subst \,/,$(shell $(GO_TOOL) -n go.opentelemetry.io/collector/cmd/mdatagen))

MDATAGEN_METADATA_YAML?= metadata.yaml

.PHONY: mdatagen
mdatagen:
	@$(MDATAGEN) $(MDATAGEN_METADATA_YAML)

.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: bench
bench:
	go test -bench=. -benchmem