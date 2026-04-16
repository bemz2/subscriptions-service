FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/bin/subscriptions-api ./cmd/subscriptions-api

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/bin/subscriptions-api .
COPY migrations ./migrations

EXPOSE 1323

CMD ["./subscriptions-api"]
