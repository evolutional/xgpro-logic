BINARY_NAME=build/xgpro-logic

all: build

deps:
	go mod tidy

build: deps
	go build -o ${BINARY_NAME} ./cmd/xgpro-logic.app/main.go

clean:
	go clean
	rm ${BINARY_NAME}