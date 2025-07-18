.PHONY: build-app
build-app:
	#CC=x86_64-unknown-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o rssgram cmd/main.go
	go build -o $(BINPATH) cmd/main.go

.PHONY: run-app
run-app:
	go run cmd/main.go

.PHONY: test
test:
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-short
test-short:
	go test -v -short ./...

.PHONY: test-integration
test-integration:
	go test -v -tags=integration ./...

.PHONY: test-benchmark
test-benchmark:
	go test -v -bench=. -benchmem ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean
clean:
	rm -f coverage.out coverage.html
	go clean -testcache