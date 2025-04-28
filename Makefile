docker-build:
	docker build -t payterbank:latest .

local:
	docker run --rm --name payterbank -p 2025:2025 -e DSN=postgresql://postgres:Sj6AR5rHTLCCJMgkJ66Rxgwh6hAkuRyXHYrr@localhost:5432/jolloy?sslmode=disable payterbank:latest

install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

swagger:
	swag init -g server/cmd/main.go
	@echo "Swagger docs generated in ./docs/swagger.json"

mock-generate:
	go generate ./...