.DEFAULT_GOAL := help

GO ?= go

GO_PACKAGES := ./...

.PHONY: help
help:
	@echo "Targets:"
	@echo "  fmt     gofmt -w . (write formatting)"
	@echo "  fmt-check  fail if any file is not gofmt-formatted"
	@echo "  vet     go vet ./..."
	@echo "  test    go test ./..."
	@echo "  tidy    go mod tidy"
	@echo "  lint    fmt-check + vet + test"

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: fmt-check
fmt-check:
	@files="$$(gofmt -l .)" || exit 1; \
	if [ -n "$$files" ]; then \
		echo "Not gofmt-formatted:"; \
		echo "$$files"; \
		exit 1; \
	fi

.PHONY: vet
vet:
	$(GO) vet $(GO_PACKAGES)

.PHONY: test
test:
	$(GO) test $(GO_PACKAGES)

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: lint
lint: fmt-check vet test
