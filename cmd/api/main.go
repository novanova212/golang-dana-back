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

	if err := db.AutoMigrate(
		&model.User{},
		&model.Bill{},
		&model.BillParticipant{},
		&model.Transaction{},
		&model.MoneyRequest{},
	); err != nil {
		log.Fatal("Gagal migrate database: ", err)
	}
	log.Println("Migrasi database berhasil")

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

	router := gin.Default()

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
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		api.POST("/wallet/topup", walletHandler.TopUp)
		api.POST("/wallet/transfer", walletHandler.Transfer)
		api.GET("/wallet/history/:user_id", walletHandler.GetHistory)

		api.POST("/bills", billHandler.CreateBill)
		api.POST("/bills/custom", billHandler.CreateCustomBill)
		api.GET("/bills/:id", billHandler.GetBillDetail)
		api.POST("/bills/participants/:participant_id/settle", billHandler.SettleParticipant)

		api.POST("/requests", reqHandler.CreateRequest)
		api.GET("/requests/incoming/:user_id", reqHandler.GetIncoming)
		api.GET("/requests/outgoing/:user_id", reqHandler.GetOutgoing)
		api.POST("/requests/:id/pay", reqHandler.PayRequest)
		api.POST("/requests/:id/decline", reqHandler.DeclineRequest)

		// Route TERLINDUNGI - butuh token JWT valid. Dipakai khusus untuk
		// operasi yang berhubungan langsung dengan akun sendiri (edit
		// profil, ganti password), supaya user_id diambil dari token,
		// bukan dari body yang bisa dipalsukan.
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/me", userHandler.GetProfile)
			protected.PUT("/me", userHandler.UpdateProfile)
			protected.PUT("/me/password", userHandler.ChangePassword)
		}
	}

	log.Println("Server berjalan di port", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal("Gagal menjalankan server: ", err)
	}
}