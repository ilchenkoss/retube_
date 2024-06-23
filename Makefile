deps:
	@go get ./...
server: deps
	@go build -o server ./cmd
test: deps
	@go test -v ./internal/core/service