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
#
#Pizda
#
pizda:
	go run ./src/pizda/main.go
pizda_build:
	 @go build -o ./tmp/pizda ./src/pizda/main.go
pizda_restart:
	@git pull
	@go build -o ./tmp/pizda ./src/pizda/main.go
	@echo Build ends
	@sudo supervisorctl restart pizda
	@echo Started
