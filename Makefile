.SILENT:
.PHONY:
.DEFAULT_GOAL := run

build:
	go build -o ./build/app cmd/bot/main.go
run:
	./scripts/run.sh

unit-test:
	go test -count=2 -short ./...

ci: unit-test
