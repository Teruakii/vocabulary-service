# เปลี่ยนจาก golang:1.22-alpine เป็น golang:alpine (หรือ golang:1.26-alpine หากใช้ 1.26)
FROM golang:alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
EXPOSE 8085
CMD ["./main"]