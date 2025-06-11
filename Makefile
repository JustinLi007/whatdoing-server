app := whatdoing

run: build
	@./whatdoing

build:
	@go build -o $(app) ./cmd/app/main.go

dockerup:
	@docker compose up -d

dockerdown:
	@docker compose down
