package middleware

// Rate limiter sederhana berbasis IP, tanpa dependency eksternal.
// Konsepnya: simpan daftar "waktu request" tiap IP dalam memori, lalu
// setiap ada request baru, buang catatan yang sudah lebih dari 1 menit,
// dan tolak kalau sisa catatan dalam 1 menit terakhir sudah >= batas.
//
// Catatan: karena disimpan di memori (bukan Redis/database), rate limit
// ini akan reset kalau server di-restart, dan tidak akan konsisten kalau
// aplikasi dijalankan di banyak instance sekaligus (perlu Redis untuk itu).
// Untuk skala proyek belajar/portofolio ini sudah cukup memadai.

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// allow mengembalikan true kalau IP ini masih boleh melakukan request.
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Buang catatan waktu yang sudah kadaluarsa (di luar jendela waktu).
	valid := rl.requests[ip][:0]
	for _, t := range rl.requests[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[ip] = valid
		return false
	}

	rl.requests[ip] = append(valid, now)
	return true
}

// RateLimitMiddleware membatasi jumlah request per IP dalam jendela waktu
// tertentu. Dipakai khusus di endpoint sensitif seperti /login untuk
// mencegah brute-force (menebak password berkali-kali secara otomatis).
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := newRateLimiter(limit, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak percobaan, coba lagi dalam beberapa saat",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}