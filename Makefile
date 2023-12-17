.PHONY: lint
lint:
	docker run -t --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.54.2 golangci-lint run -v --timeout=200s

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/main ./cmd/main.go

.PHONY: run
run:
	go run ./microservices/user/cmd/server.go &
	go run ./microservices/auth/cmd/server.go &
	go run ./microservices/passage/cmd/server.go &
	go run ./cmd/main.go &

.PHONY: kill
kill:
	fuser -k 8080/tcp || true
	fuser -k 8081/tcp || true
	fuser -k 8082/tcp || true
	fuser -k 8083/tcp || true

.PHONY: deploy
run:
	/usr/local/go run ./microservices/user/cmd/server.go &
	/usr/local/go run ./microservices/auth/cmd/server.go &
	/usr/local/go run ./microservices/passage/cmd/server.go &
	/usr/local/go run ./cmd/main.go &

.PHONY: test
test:
	go test ./... -coverprofile cover.out.tmp && cat cover.out.tmp > ./cover.out && rm cover.out.tmp && go tool cover -func ./cover.out

.PHONY: coverage
coverage:
	go tool cover -html ./coverage/cover.out
