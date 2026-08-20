.PHONY: start test-get test-post test-nc

start:
	go run ./main.go

test-get:
	nc localhost 8080 < requests/get.txt

test-post:
	nc localhost 8080 < requests/post.txt
