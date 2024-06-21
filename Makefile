deps:
	@go get ./...
server: deps
	@go build -o server ./cmd