package config

// Package ini mirip fungsi config/database.php + .env di Laravel.
// Tugasnya: baca environment variable, lalu bikin koneksi ke database.

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Struct ini nampung semua config yang dibaca dari .env
// Mirip kayak kamu akses env('DB_HOST') dkk di Laravel, tapi di sini
// kita kumpulin jadi satu struct biar rapi dan gampang di-pass ke fungsi lain.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	AppPort    string
}

// LoadConfig membaca file .env dan mengembalikan struct Config.
// Dipanggil sekali di awal (main.go), mirip bootstrap Laravel baca .env.
func LoadConfig() *Config {
	// godotenv.Load() cari file .env di root project.
	// Kalau tidak ketemu, kita cuma warning (bukan fatal error),
	// karena di production biasanya env variable di-set langsung di server,
	// bukan lewat file .env.
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: file .env tidak ditemukan, menggunakan environment variable sistem")
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "dana_clone"),
		JWTSecret:  getEnv("JWT_SECRET", "secret"),
		AppPort:    getEnv("APP_PORT", "8080"),
	}
}

// getEnv adalah helper: ambil environment variable, kalau kosong pakai default.
// Mirip fungsi env('KEY', 'default') di Laravel.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// ConnectDB membuka koneksi ke PostgreSQL menggunakan GORM.
// Ini mirip apa yang terjadi otomatis di balik layar saat Laravel
// baca config/database.php dan bikin koneksi PDO ke database.
func ConnectDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// log.Fatal akan menghentikan program kalau koneksi gagal.
		// Wajar dilakukan di awal startup: kalau DB tidak bisa konek,
		// tidak ada gunanya lanjut menjalankan server.
		log.Fatal("Gagal konek ke database: ", err)
	}

	log.Println("Berhasil konek ke database:", cfg.DBName)
	return db
}
