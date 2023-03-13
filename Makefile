.SILENT:
.PHONY:
.DEFAULT_GOAL := run

run:
	go build -o ./build/app cmd/main.go && xargs < .env ./build/app -strict=false -config-path=./config.yml

unit-test:
	go test -count=5 -short ./...

ci: unit-test
