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
deploy:
	nohup /usr/local/go/bin/go run ./microservices/user/cmd/server.go > user.out 2> user.err < /dev/null &
	nohup /usr/local/go/bin/go run ./microservices/auth/cmd/server.go > auth.out 2> auth.err < /dev/null &
	nohup /usr/local/go/bin/go run ./microservices/passage/cmd/server.go > passage.out 2> passage.err < /dev/null &
	nohup /usr/local/go/bin/go run ./cmd/main.go > main.out 2> main.err < /dev/null &

.PHONY: test
test:
	go test ./... -coverprofile cover.out.tmp && cat cover.out.tmp > ./cover.out && rm cover.out.tmp && go tool cover -func ./cover.out

.PHONY: coverage
coverage:
	go tool cover -html ./coverage/cover.out
