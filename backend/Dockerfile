FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./
# The "./..." tells go to build every file that has a main function.
RUN swag init && mkdir -p bin/ && go build -o bin/ ./...
# We can then copy the different binary files into indivdual containers below.

# All-in-one image; for testing purposes
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


FROM alpine:latest AS board

WORKDIR /app
COPY --from=builder /workspace/bin/board /app/board

EXPOSE 80

CMD ["/app/board"]


FROM alpine:latest AS card

WORKDIR /app
COPY --from=builder /workspace/bin/card /app/card

EXPOSE 80

CMD ["/app/card"]


FROM alpine:latest AS column

WORKDIR /app
COPY --from=builder /workspace/bin/column /app/column

EXPOSE 80

CMD ["/app/column"]


FROM alpine:latest AS docs

WORKDIR /app
COPY --from=builder /workspace/bin/docs /app/docs

EXPOSE 80

CMD ["/app/docs"]


FROM alpine:latest AS overview

WORKDIR /app
COPY --from=builder /workspace/bin/overview /app/overview

EXPOSE 80

CMD ["/app/overview"]
