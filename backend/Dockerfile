FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./
RUN swag init && mkdir -p bin/ && go build -o bin/ ./...


FROM alpine:latest AS aio

WORKDIR /app
COPY --from=builder /workspace/bin/cwm25 /app/server

EXPOSE 80

CMD ["/app/server"]


FROM alpine:latest AS auth

WORKDIR /app
COPY --from=builder /workspace/bin/auth /app/auth

EXPOSE 80

CMD ["/app/auth"]
