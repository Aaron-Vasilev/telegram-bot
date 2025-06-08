run:
	 @go run main.go
build:
	 @go build -o ./tmp/bot ./main.go
format:
	@gofmt -w .
start:
	@supervisord -c ./supervisord.conf
restart:
	@go build -o ./tmp/bot ./main.go
	@echo Build ends
	@sudo supervisord restart bot
	@echo Started
stop:
	@supervisorctl shutdown
