COVERAGE_FILE := "./cover.out"
# Run build
build:
	go build ./...

tidy:
	go mod tidy

# display intra-package 
dep:
	go list -f '{{ join .Imports "\n" }}' ./$(PKG) | grep 'github.com/tradeline-tech/workflow/'

# Run linter
lint:
	golangci-lint run

# Run all tests
test:
	go test ./...

# Run all tests and get coverage report
coverage:
	go test ./... -coverprofile $(COVERAGE_FILE) && \
	rm $(COVERAGE_FILE)

# Run all tests and get HTML coverage report
coverage-html:
	go test ./... -coverprofile $(COVERAGE_FILE) && \
	go tool cover -html=$(COVERAGE_FILE) && \
	rm $(COVERAGE_FILE)
