package main

import (
	"log"
	"net/http"

	"dana-clone/internal/config"
	"dana-clone/internal/handler"
	"dana-clone/internal/middleware"
	"dana-clone/internal/model"
	"dana-clone/internal/repository"
	"dana-clone/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)

	if err := db.AutoMigrate(&model.User{}, &model.Bill{}, &model.BillParticipant{}); err != nil {
		log.Fatal("Gagal migrate database: ", err)
	}
	log.Println("Migrasi database berhasil")

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userRepo)
	walletService := service.NewWalletService(userRepo, db)
	walletHandler := handler.NewWalletHandler(walletService)
	billRepo := repository.NewBillRepository(db)
	billService := service.NewBillService(billRepo, walletService)
	billHandler := handler.NewBillHandler(billService)

	router := gin.Default()

	// CORS middleware: mengizinkan request dari browser (file lokal atau
	// domain lain) untuk mengakses API ini. Tanpa ini, browser akan
	// memblokir request dari frontend HTML/Vue ke API Go ini.
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Browser kadang mengirim request "OPTIONS" dulu sebagai
		// pengecekan awal (preflight) sebelum request asli dikirim.
		// Kita cukup balas 204 (No Content) untuk request semacam ini.
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	api := router.Group("/api")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.POST("/wallet/topup", walletHandler.TopUp)
		api.POST("/wallet/transfer", walletHandler.Transfer)
		api.POST("/bills", billHandler.CreateBill)
		api.GET("/bills/:id", billHandler.GetBillDetail)
		api.POST("/bills/participants/:participant_id/settle", billHandler.SettleParticipant)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/me", userHandler.GetProfile)
		}
	}

	log.Println("Server berjalan di port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal("Gagal menjalankan server: ", err)
	}
}