ifneq (,$(wildcard ./.env))
    include .env
    export
endif

build: gen
	go build -o ./bin/app.exe ./cmd/app

run: build
	./bin/app -local

PB=pb
PROTO=proto/auth.proto
protobuf:
	protoc  \
		--go_out=. \
		--go_opt=M$(PROTO)=$(PB)/authpb \
		--go-grpc_out=. \
		--go-grpc_opt=M$(PROTO)=$(PB)/authpb \
		$(PROTO)

wire-gen:
	wire ./internal/app

swagger:
	swag init -g .\internal\transport\http\http.go

gen: wire-gen swagger

coverage:
	go test -v -coverprofile cover.out ./...
	go tool cover -html cover.out -o cover.html
	rm cover.out

migrate.up:
	migrate -path ./migrations -database 'postgres://$(PG_USER):$(PG_PASS)@$(PG_HOST):$(PG_PORT)/$(PG_NAME)?sslmode=disable' up

migrate.down:
	migrate -path ./migrations -database 'postgres://$(PG_USER):$(PG_PASS)@$(PG_HOST):$(PG_PORT)/$(PG_NAME)?sslmode=disable' down