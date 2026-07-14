include .env

migrate-up:
	migrate -path migrations -database $(DATABASE_URL) up

migrate-down:
	migrate -path migrations -database $(DATABASE_URL) down 1

psql:
	docker exec -it expense-tracker-db psql -U ${DATABASE_USER} -d ${DATABASE_NAME}

run:
	@go run .
