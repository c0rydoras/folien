make:
	go run main.go examples/folien.md

test:
	go test ./... -short

build:
	go build -o folien
