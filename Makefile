MODULE_NAME=unisys
BIN_DIR=bin
CONF_DIR=conf
AUTH_DIR=auth
CONF_FILE=unisys.yaml
BUILD_TIME=$(shell date +%Y-%m-%d' '%H:%M:%S)

define go_build
	mkdir -p ${BIN_DIR}/${CONF_DIR}
	go build -o ${BIN_DIR}/${MODULE_NAME} -ldflags "-X 'config.BuildTime=${BUILD_TIME}'"
	cp -f config/${CONF_FILE} ${BIN_DIR}/${CONF_DIR}/${CONF_FILE}
	# TLS 인증서 디렉터리 복사 (테스트용)
	# cp -rf ${AUTH_DIR} ${BIN_DIR}
endef

all: init build

init:
	@if [ ! -f go.mod ]; then \
		echo "Initialize Go Module..."; \
		go mod init github.com/meloncoffee/${MODULE_NAME}; \
		go mod tidy; \
	fi
	
deps:
	@if [ -f go.mod ]; then \
		echo "Installing Dependencies..."; \
		go mod tidy; \
	fi

build:
	@echo "Building Project..."
	$(call go_build)

clean:
	@echo "Cleaning up..."
	rm -rf ${BIN_DIR}

.PHONY: init deps build clean
