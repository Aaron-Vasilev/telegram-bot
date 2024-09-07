run:
	 @go run main.go
build:
	 @go build -o ./build/bot ./main.go
format:
	@gofmt -w .
restart:
	@supervisorctl shutdown
	@go build -o ./build/bot ./main.go
	@supervisord -c ./supervisord.conf
stop:
	@supervisorctl shutdown
