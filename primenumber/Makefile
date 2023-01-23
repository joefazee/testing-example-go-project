COVERAGE=coverage.out


test:
	@go test -v ./...


cover:
	@go test -coverprofile=$COVERAGE
	@go tool cover -html=$COVERAGE

clean:
	@rm -f $COVERAGE

run:
	@go run .