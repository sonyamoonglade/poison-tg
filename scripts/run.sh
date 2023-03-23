export $(xargs < .env)
go build -o ./build/app cmd/bot/main.go && \
./build/app -strict=false -config-path=./config.yml
