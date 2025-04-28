FROM golang:1.23.8 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o payterbank /app/server/cmd/main.go

FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /app/docs       /app/docs
COPY --from=builder /app/payterbank /app/payterbank
COPY --from=builder /app/migrations /app/migrations
EXPOSE 2025
ENTRYPOINT ["/app/payterbank"]