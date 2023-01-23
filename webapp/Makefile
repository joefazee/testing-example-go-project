COVERAGE=coverage.out


test:
	@go test -v ./...

test-integration:
	@go test -v -tags=integration ./...

cover:
	@go test ./... -coverprofile=$COVERAGE
	@go tool cover -html=$COVERAGE

clean:
	@rm -f $COVERAGE

run-web:
	@go run ./cmd/web

run-api:
	@go run ./cmd/api

docker-up:
	@docker compose up