
.PHONY: run run-debug test

run: # Run the service on localhost.
	go run ./...

run-debug: # Run the service with debug mode enabled.
	go run ./... -debug

test: # Run unit tests.
	go test ./... -v --cover
