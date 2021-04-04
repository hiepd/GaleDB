APP_EXECUTABLE=galedb

compile:
	mkdir -p out
	go build -o out/${APP_EXECUTABLE} cmd/main.go

run:
	./out/${APP_EXECUTABLE}