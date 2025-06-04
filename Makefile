app := whatdoing

run: build
	@./whatdoing

build:
	@go build -o $(app) ./cmd/whatdoing/main.go
