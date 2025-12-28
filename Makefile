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
load_test:
	@k6 run test/load_test.js --out web-dashboard=export=report.html
