# Stage 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# 1. Arahkan ke file main.go yang spesifik
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

# Stage 2: Final Image
FROM alpine:latest
WORKDIR /root/

# 2. Copy binary 'main' yang sudah di-build dari Stage 1
COPY --from=builder /app/main .

EXPOSE 8080

# 3. Jalankan binary-nya langsung (karena lokasinya sekarang di /root/main)
CMD ["./main"]