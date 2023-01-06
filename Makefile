build:
	GOARCH=amd64 GOOS=linux go build -o dist/api

build-mac:
	go build -o dist/api-mac

run:
	go run main.go

deploy: build
	sls deploy --config serverless.yml
