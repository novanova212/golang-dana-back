package middleware

// Structured logging: setiap log dicatat dalam format JSON dengan
// field-field yang konsisten (method, path, status, durasi, dst),
// bukan sekadar teks bebas seperti log.Println biasa.
//
// Ini penting untuk aplikasi production karena log JSON bisa dengan
// mudah di-parse, difilter, dan dicari oleh tools log management
// (misal Datadog, ELK stack, Grafana Loki), sesuatu yang sulit dilakukan
// kalau formatnya teks bebas tak beraturan.
//
// log/slog adalah package bawaan standard library Go (sejak Go 1.21),
// jadi tidak perlu dependency eksternal tambahan.

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// StructuredLogger mencatat setiap request yang masuk dengan format JSON.
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next() // proses request dulu, baru catat log setelah selesai

		duration := time.Since(start)
		status := c.Writer.Status()

		logLevel := slog.LevelInfo
		if status >= 500 {
			logLevel = slog.LevelError
		} else if status >= 400 {
			logLevel = slog.LevelWarn
		}

		slog.Log(c.Request.Context(), logLevel, "http_request",
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}