run:
	 @go run main.go
build:
	 @go build -o ./tmp/bot ./main.go
format:
	@gofmt -w .
start:
	@supervisord -c ./supervisord.conf
restart:
	@sudo supervisorctl shutdown
	@go build -o ./tmp/bot ./main.go
	@echo Build ends
	@sudo supervisord -c ./supervisord.conf
	@echo Started
stop:
	@supervisorctl shutdown
