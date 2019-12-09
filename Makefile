RUNNER = "bin/tgj-bot"
DB_MIGRATIONS = "db_migrations/sql"

build:
	go build -o $(RUNNER) ./cmd

run: build
	$(RUNNER) ./conf/conf.json

run_test: build
	$(RUNNER) ./conf/test_conf/conf.json

create_migration:
	@read -p "Enter migration name: " migration_name; \
	migrate create -ext sql -dir $(DB_MIGRATIONS) -seq $${migration_name}

migrate:
	@echo "RUN CMD MANUALLY:\n\
	migrate -database postgres://postgres:password@localhost:5432/bot?sslmode=disable -path $(DB_MIGRATIONS) up\n\
	migrate -database postgres://postgres:password@localhost:5432/bot?sslmode=disable -path $(DB_MIGRATIONS) down"

init_tests:
	go run external_service/database/tests/prepare.go ./conf/conf.json

tests: init_tests
	find ./ -name '*_test.go' | sort | xargs -n1 dirname | xargs go test -count=1 -cover
