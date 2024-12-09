.PHONY: test clean
default: build

BINARY_FILE_NAME=pub
COVERAGE_FILE_NAME=cover.out
COVERAGE_TMP_FILE_NAME=cover.tmp

proto:
	go install github.com/golang/protobuf/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	PATH=${PATH}:~/go/bin protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative \
		api/grpc/publisher/*.proto \
		api/grpc/subject/*.proto \
		api/grpc/tgbot/*.proto \
		api/grpc/limits/*.proto \
		api/grpc/permits/*.proto \
		api/grpc/source/*/*.proto \
		api/grpc/events/*.proto \
		api/grpc/auth/*.proto \
		api/grpc/ce/*.proto

vet: proto
	go vet

test: vet
	go test -race -cover -coverprofile=${COVERAGE_FILE_NAME} ./...
	cat ${COVERAGE_FILE_NAME} | grep -v _mock.go | grep -v logging.go | grep -v .pb.go > ${COVERAGE_FILE_NAME}.tmp
	mv -f ${COVERAGE_FILE_NAME}.tmp ${COVERAGE_FILE_NAME}
	go tool cover -func=${COVERAGE_FILE_NAME} | grep -Po '^total\:\h+\(statements\)\h+\K.+(?=\.\d+%)' > ${COVERAGE_TMP_FILE_NAME}
	./scripts/cover.sh
	rm -f ${COVERAGE_TMP_FILE_NAME}

build: proto
	CGO_ENABLED=0 GOOS=linux GOARCH= GOARM= go build -ldflags="-s -w" -o ${BINARY_FILE_NAME} main.go
	chmod ugo+x ${BINARY_FILE_NAME}

docker:
	docker build -t awakari/pub .

run: docker
	docker run \
		-d \
		--name awakari-pub \
		-p 8080:8080 \
		--expose 8080 \
		awakari/pub

staging: docker
	./scripts/staging.sh

release: docker
	./scripts/release.sh

clean:
	go clean
	rm -f ${BINARY_FILE_NAME} ${COVERAGE_FILE_NAME}
