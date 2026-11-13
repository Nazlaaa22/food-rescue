
# Menggunakan Go sebagai base image
FROM golang:1.26-alpine AS builder

# Menentukan direktori kerja
WORKDIR /app

# Menyalin file dependency
COPY go.mod go.sum ./

# Mengunduh dependency
RUN go mod download

# Menyalin seluruh source code
COPY . .

# Build aplikasi
RUN CGO_ENABLED=0 GOOS=linux go build -o food-rescue ./cmd/api

# Menggunakan image ringan untuk menjalankan aplikasi
FROM alpine:latest

WORKDIR /app

# Menyalin binary dari builder
COPY --from=builder /app/food-rescue .

# Membuka port aplikasi
EXPOSE 8080

# Menjalankan aplikasi
CMD ["./food-rescue"]
