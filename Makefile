.PHONY:vet lint upgrade
vet:
	@echo "Running go vet"
	@go mod tidy
	@go vet ./...
lint:vet
	@echo "Running golangci-lint"
	@hash golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run
upgrade:
	@echo "Upgrading..."
	@go mod tidy
	@go get -d -u -v ./...
	@go mod tidy
	@echo "Upgrade complete."