.PHONY: dev-api dev-web test

dev-api:
	cd apps/api && go run ./cmd/server

dev-web:
	cd apps/web && npm run dev

test:
	cd apps/api && go test ./...
	cd apps/web && npm run build
