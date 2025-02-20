.PHONY: docs

VERSION="0.0.1"

dev: docs
	go run cmd/restful/main.go

migrate-create:
	@read -p  "Migration name (eg:create_users, alter_entities, ...): " NAME; \
	migrate create -ext sql -seq -dir database/migrations $$NAME

migrate-up:
	migrate -database ${DATABASE_URL} -path database/migrations up

migrate-down:
	migrate -database ${DATABASE_URL} -path database/migrations down -all

migrate-status:
	migrate -database ${DATABASE_URL} -path database/migrations version

migrate-clean:
	@read -p  "Force to version: " VERSION; \
	migrate -database ${DATABASE_URL} -path database/migrations force $$VERSION

docs:
	swag fmt
	swag init -g cmd/restful/main.go