.PHONY: all

migrate-create:
	@read -p  "Migration name (eg:create_users, alter_entities, ...): " NAME; \
	migrate create -ext sql -seq -dir database/migrations $$NAME

migrate-up:
	migrate -database ${DATABASE_URL} -path database/migrations up

migrate-down:
	migrate -database ${DATABASE_URL} -path database/migrations down -all
