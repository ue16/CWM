FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./
RUN swag init && go build -o /server .


FROM alpine:latest

WORKDIR /app
COPY --from=builder /server /app/server
COPY static/* /app/static/
COPY templates/* /app/templates/

EXPOSE 80

CMD ["/app/server"]
