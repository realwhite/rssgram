
.PHONY: build-app
build-app:
	go build -o $(BINPATH) cmd/main.go


.PHONY: run-app
run-app:
	go run cmd/main.go