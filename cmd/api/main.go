package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"dana-clone/internal/config"
	"dana-clone/internal/handler"
	"dana-clone/internal/middleware"
	"dana-clone/internal/model"
	"dana-clone/internal/repository"
	"dana-clone/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Setup structured logger sebagai default logger aplikasi.
	// Semua log (termasuk dari log/slog di file lain) otomatis
	// mengikuti format JSON ini.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)

	if err := db.AutoMigrate(
		&model.User{},
		&model.Bill{},
		&model.BillParticipant{},
		&model.Transaction{},
		&model.MoneyRequest{},
	); err != nil {
		slog.Error("gagal migrate database", "error", err)
		os.Exit(1)
	}
	slog.Info("migrasi database berhasil")

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userRepo, authService)

	txRepo := repository.NewTransactionRepository(db)
	walletService := service.NewWalletService(userRepo, txRepo, db)
	walletHandler := handler.NewWalletHandler(walletService)

	billRepo := repository.NewBillRepository(db)
	billService := service.NewBillService(billRepo, walletService)
	billHandler := handler.NewBillHandler(billService)

	reqRepo := repository.NewMoneyRequestRepository(db)
	reqService := service.NewMoneyRequestService(reqRepo, walletService)
	reqHandler := handler.NewMoneyRequestHandler(reqService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.StructuredLogger())

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	api := router.Group("/api")
	{
		// ===== PUBLIK (tidak butuh login) =====
		api.POST("/register", authHandler.Register)
		// Login dibatasi rate limit 10x per menit per IP, mencegah
		// brute-force menebak password secara otomatis.
		api.POST("/login", middleware.RateLimitMiddleware(10, time.Minute), authHandler.Login)
		// Detail bill boleh dilihat siapa saja yang punya link/ID-nya,
		// mirip membuka invoice - tidak mengubah data apa pun.
		api.GET("/bills/:id", billHandler.GetBillDetail)

		// ===== PROTECTED (wajib token JWT valid) =====
		// Semua operasi yang berhubungan dengan identitas/akun sendiri
		// (top up saldo sendiri, transfer dari saldo sendiri, dst) WAJIB
		// lewat sini, supaya user_id diambil dari token yang terverifikasi,
		// bukan dari body request yang bisa dipalsukan oleh siapa saja.
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/me", userHandler.GetProfile)
			protected.PUT("/me", userHandler.UpdateProfile)
			protected.PUT("/me/password", userHandler.ChangePassword)

			protected.POST("/wallet/topup", walletHandler.TopUp)
			protected.POST("/wallet/transfer", walletHandler.Transfer)
			protected.GET("/wallet/history", walletHandler.GetHistory)

			protected.POST("/bills", billHandler.CreateBill)
			protected.POST("/bills/custom", billHandler.CreateCustomBill)
			protected.POST("/bills/participants/:participant_id/settle", billHandler.SettleParticipant)

			protected.POST("/requests", reqHandler.CreateRequest)
			protected.GET("/requests/incoming", reqHandler.GetIncoming)
			protected.GET("/requests/outgoing", reqHandler.GetOutgoing)
			protected.POST("/requests/:id/pay", reqHandler.PayRequest)
			protected.POST("/requests/:id/decline", reqHandler.DeclineRequest)
		}
	}

	slog.Info("server berjalan", "port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		slog.Error("gagal menjalankan server", "error", err)
		os.Exit(1)
	}
}