# Dana Clone — Backend API

Aplikasi e-wallet sederhana (terinspirasi DANA/GoPay) yang dibangun dengan Go, sebagai proyek belajar backend development dari nol hingga siap dipakai portofolio.

## ✨ Fitur

- **Autentikasi** — Register, Login, JWT token, password ter-hash (bcrypt)
- **Wallet** — Top up saldo, transfer antar user (aman dari race condition lewat database transaction + row locking)
- **Riwayat Transaksi** — Tercatat otomatis untuk setiap top up/transfer, bisa difilter berdasarkan tipe dan dicari berdasarkan keterangan
- **Split Bill** — Bagi tagihan rata atau custom per orang, lengkap dengan pelunasan (settle) yang memicu transfer saldo beneran
- **Minta Uang (Money Request)** — Tagih teman, penerima bisa bayar atau tolak
- **Edit Profil & Ganti Password**

## 🛠️ Tech Stack

- **Go** — bahasa utama
- **Gin** — HTTP web framework
- **GORM** — ORM untuk akses PostgreSQL
- **PostgreSQL** — database
- **JWT** (golang-jwt) — autentikasi berbasis token
- **bcrypt** — hashing password

## 🏗️ Arsitektur

Proyek ini mengikuti pola **Clean Architecture** sederhana, memisahkan tanggung jawab tiap layer:

```
Request → Handler → Service → Repository → Database
```

- **handler** — menerima HTTP request, validasi input dasar, kembalikan response
- **service** — logika bisnis (validasi aturan, kalkulasi, orkestrasi)
- **repository** — satu-satunya layer yang bicara langsung ke database
- **model** — representasi struktur data / tabel

Struktur folder:
```
danaclone-back/
├── cmd/api/main.go          # entry point, wiring dependency & routing
├── internal/
│   ├── config/               # load .env, koneksi database
│   ├── model/                 # struct User, Bill, Transaction, dll
│   ├── repository/            # akses database
│   ├── service/                 # logika bisnis
│   ├── handler/                 # HTTP handler (mirip Controller)
│   └── middleware/               # JWT auth middleware
├── go.mod
└── .env.example
```

## 🚀 Cara Menjalankan

### Prasyarat
- Go 1.21+
- PostgreSQL

### Setup

1. Clone repo ini
2. Buat database PostgreSQL:
   ```sql
   CREATE DATABASE dana_clone;
   ```
3. Copy `.env.example` jadi `.env`, sesuaikan kredensial database:
   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your_password
   DB_NAME=dana_clone
   JWT_SECRET=ganti-dengan-string-rahasia
   APP_PORT=8080
   ```
4. Install dependency:
   ```
   go mod download
   ```
5. Jalankan:
   ```
   go run cmd/api/main.go
   ```

Server akan jalan di `http://localhost:8080`, dan tabel database otomatis dibuat (auto-migrate) saat pertama kali dijalankan.

## 📡 API Endpoints

### Auth
| Method | Endpoint | Keterangan |
|---|---|---|
| POST | `/api/register` | Daftar akun baru |
| POST | `/api/login` | Login, mengembalikan JWT token |

### Profil (perlu token)
| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/api/me` | Lihat profil & saldo sendiri |
| PUT | `/api/me` | Update nama |
| PUT | `/api/me/password` | Ganti password |

### Wallet
| Method | Endpoint | Keterangan |
|---|---|---|
| POST | `/api/wallet/topup` | Top up saldo |
| POST | `/api/wallet/transfer` | Transfer ke user lain |
| GET | `/api/wallet/history/:user_id?type=&search=` | Riwayat transaksi (filter opsional) |

### Split Bill
| Method | Endpoint | Keterangan |
|---|---|---|
| POST | `/api/bills` | Buat bill, bagi rata |
| POST | `/api/bills/custom` | Buat bill, porsi custom per orang |
| GET | `/api/bills/:id` | Detail bill + status pelunasan tiap peserta |
| POST | `/api/bills/participants/:participant_id/settle` | Lunasi porsi sendiri |

### Minta Uang
| Method | Endpoint | Keterangan |
|---|---|---|
| POST | `/api/requests` | Buat permintaan uang |
| GET | `/api/requests/incoming/:user_id` | Permintaan masuk (kamu ditagih) |
| GET | `/api/requests/outgoing/:user_id` | Permintaan keluar (kamu menagih) |
| POST | `/api/requests/:id/pay` | Bayar permintaan |
| POST | `/api/requests/:id/decline` | Tolak permintaan |

## 🔒 Keamanan yang Diterapkan

- Password di-hash dengan bcrypt, tidak pernah disimpan/dikirim dalam bentuk plain text
- JWT dengan masa berlaku (expiry) 24 jam
- Endpoint yang mengubah data sensitif milik sendiri (edit profil, ganti password) mengambil `user_id` dari token JWT (bukan dari body request), mencegah orang lain mengedit akun orang lain
- Transfer saldo menggunakan **database transaction + row locking** (`SELECT ... FOR UPDATE`) untuk mencegah race condition saat banyak request bersamaan

## 📝 Catatan Pengembangan

Proyek ini dibangun sebagai media belajar Go dari nol, dengan progres bertahap:
1. Fundamental Go & struktur clean architecture
2. Autentikasi & JWT
3. Transaksi database & pencegahan race condition
4. Fitur bisnis (split bill, money request)
5. Integrasi dengan frontend Vue.js

Fitur yang bisa dikembangkan selanjutnya: notifikasi real-time, QR code payment, laporan bulanan per kategori, unit testing otomatis.