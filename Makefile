run:
	 @go run ./src/main.go
build:
	 @go build -o ./tmp/bot ./src/main.go
format:
	@gofmt -w .
start:
	@supervisord -c ./supervisord.conf
restart:
	@go build -o ./tmp/bot ./src/main.go
	@echo Build ends
	@sudo supervisorctl restart bot
	@echo Started
stop:
	@supervisorctl shutdown
sql:
	@sqlc generate
