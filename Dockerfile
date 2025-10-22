FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/server/main.go

FROM alpine:latest

WORKDIR /root/

# Copy compiled binary from builder
COPY --from=builder /app/server .
COPY db/init.sql ./db/init.sql

EXPOSE 8080

CMD ["./server"]
