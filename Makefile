APP_NAME=proxy_checker
BUILD_DIR=../bin

.PHONY: build run clean

build:
	cd proxy_checker && go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/main.go

run: build
	cd proxy_checker && CONFIG_PATH=config.yml $(BUILD_DIR)/$(APP_NAME)

clean:
	rm -rf proxy_checker/../bin
