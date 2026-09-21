# ===== Stage 1: Build =====
# Pakai image Go lengkap untuk compile, tapi image ini TIDAK dipakai
# di hasil akhir - cuma "numpang" buat proses build.
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod & go.sum dulu SEBELUM copy semua source code.
# Trik ini memanfaatkan Docker layer caching: kalau dependency tidak
# berubah, Docker tidak perlu download ulang tiap kali build - cuma
# perlu build ulang kalau source code-nya yang berubah.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 menghasilkan binary statis (tidak bergantung library C
# sistem operasi), penting supaya binary ini bisa jalan di image alpine
# yang minim di stage berikutnya.
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/api

# ===== Stage 2: Run =====
# Image based alpine yang sangat kecil (~5MB), cuma isinya OS minimal.
# Hasil akhirnya: image jauh lebih kecil dibanding kalau kita bawa
# seluruh toolchain Go (yang berat) ke production.
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]