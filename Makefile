
.PHONY: build-app
build-app:
	#CC=x86_64-unknown-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o rssgram cmd/main.go
	go build -o $(BINPATH) cmd/main.go


.PHONY: run-app
run-app:
	go run cmd/main.go