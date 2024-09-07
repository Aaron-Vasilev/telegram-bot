run:
	 @go run main.go
build:
	 @go build -o ./build/bot ./main.go
format:
	@gofmt -w .
stop:
	@supervisorctl shutdown
	@go build -o ./build/bot ./main.go
	@supervisorctl start bot
