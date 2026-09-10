# Panduan Setup

## 1. Salin file ini ke folder project kamu
Extract isi zip ini, lalu salin/timpa ke folder project Go kamu
(folder yang sudah ada `go.mod` di dalamnya, misal `dana_clone`).
Kalau ada file yang sama, timpa saja (kecuali kamu sudah edit `go.mod` sendiri).

## 2. Bikin file .env
File `.env.example` di sini isinya contoh. Salin jadi file baru bernama `.env`
(hapus `.example`-nya), lalu isi `DB_PASSWORD` dengan password PostgreSQL kamu.

Caranya lewat CMD (di dalam folder project):
```
copy .env.example .env
```
Lalu buka file `.env` pakai Notepad/VS Code, ganti `DB_PASSWORD=postgres`
dengan password asli kamu kalau berbeda.

## 3. Install semua dependency
Jalankan satu-satu di CMD (posisi di dalam folder project):
```
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/joho/godotenv
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
```

## 4. Jalankan aplikasi
```
go run cmd/api/main.go
```

Kalau berhasil, akan muncul log seperti:
```
Berhasil konek ke database: dana_clone
Migrasi database berhasil
Server berjalan di port 8080
```

## 5. Test pakai Postman / curl
Register:
```
POST http://localhost:8080/api/register
Content-Type: application/json

{
  "name": "Budi",
  "email": "budi@example.com",
  "password": "rahasia123"
}
```

Login:
```
POST http://localhost:8080/api/login
Content-Type: application/json

{
  "email": "budi@example.com",
  "password": "rahasia123"
}
```
