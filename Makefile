build:
	go build -o todo ./cmd/todo

run: build
	./todo
test:
	go test -v ./...
lint:
	golangci-lint run ./...
format:
	go fmt ./...
