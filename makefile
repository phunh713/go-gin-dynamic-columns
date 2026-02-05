include .env

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

dev:
	air

seed:
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" != "" ]; then \
		go run cmd/seed/*.go --domains $(filter-out $@,$(MAKECMDGOALS)); \
	else \
		go run cmd/seed/*.go; \
	fi

migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DB_URL)" status

# Catch-all target to allow passing domain names as arguments
%:
	@: