# Setup the platform
setup:
	@echo "Generating env files..."
	cp sample.env .env
	cp sample.env test.env
	@echo ".env & test.env created. Now, update values in them"

# Migrate database
migrate:
	@echo "Running migrations..."
	sh script/migration.sh

# Run pretest script
pretest:
	sh scripts/test_helper.sh

# Migrate Tests
migrate-test:
	@echo "Running migrations for tests..."
	sh script/migrate-tests.sh

# Run Tests
test-cover: migrate-test
	go test `go list ./... | grep -v cmd` -coverprofile=/tmp/coverage.out -coverpkg=./...
	go tool cover -html=/tmp/coverage.out

# Generate API documentation
doc:
	@echo "Generating swagger docs..."
	swag fmt --exclude ./internal/domain
	swag init --parseDependency --parseInternal -g internal/http/api/chat_api.go -ot go,yaml -o internal/http/swagger

wire:
	cd internal/dependency/ && wire && cd ../..

# Build the platform
build: migrate doc
	@echo "Building chatApp-api..."
	sh script/build.sh

# Clean the platform
clean:
	@echo "Cleaning up..."
	rm ./bin/chatApp || true
	go clean -testcache

# Stop the platform
stop:
	pkill bhoomi || true

# Start the platform
start: stop build
	nohup ./bin/chatApp &


