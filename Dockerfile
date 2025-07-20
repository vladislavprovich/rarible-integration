FROM golang:1.24.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o rarible-service ./cmd/raribleintegration

FROM alpine:latest

WORKDIR /app

RUN apk update && apk add --no-cache ca-certificates

COPY --from=builder /app/rarible-service .
COPY .env .

ENTRYPOINT ["./rarible-service"]
