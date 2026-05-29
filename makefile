.PHONY: build clean

BINARY_NAME=sfm_content

build:
	go build -ldflags="-X main.ReleaseMode=false" -o dist/${BINARY_NAME} ./

run: build
	dist/${BINARY_NAME}

prod:
	go env -w CGO_ENABLED=0
	@echo 'Building go app'
	go build -o dist/${BINARY_NAME} ./

clean:
	go clean
