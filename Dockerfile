FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/main ./cmd/main.go

FROM alpine:latest

WORKDIR /usr/bin
COPY --from=builder /app/bin/main ./

RUN chmod +x ./main

ENV HTTP_ADDR=:8080
EXPOSE 8080

ENTRYPOINT ["./main"]
