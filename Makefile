build:
	GOARCH=amd64 GOOS=linux go build -o dist/handler

build-mac:
	go build -o dist/handler-mac

run:
	go run main.go
