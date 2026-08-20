.PHONY: start test-get test-post test-post test-404

start:
	go run ./main.go

test-get:
	curl -v http://localhost:8080/

test-users:
	curl -v http://localhost:8080/users

test-post:
	curl -v -X POST http://localhost:8080/users \
		-H "Content-Type: text/plain" \
		-d "This is super horsey TCP/IP payload."

test-404:
	curl -v http://localhost:8080/non-existent-path