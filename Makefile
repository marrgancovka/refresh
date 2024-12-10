build_:
	go build -o ./.bin cmd/main.go

build:
	docker compose build

start: build
	docker compose up -d

stop:
	docker compose down

test:
	go test -race ./...

.PHONY: lint
lint:
	golangci-lint run --config=.golangci.yaml