package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"ai_ad_platform_recall_process/internal/config"
	"ai_ad_platform_recall_process/internal/handler"
	"ai_ad_platform_recall_process/internal/middleware"
	"ai_ad_platform_recall_process/internal/model"
	"ai_ad_platform_recall_process/internal/repository"
	"ai_ad_platform_recall_process/internal/service"
	"ai_ad_platform_recall_process/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件，不影响已存在的环境变量（生产环境可直接设置环境变量）
	_ = godotenv.Load()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.Init(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	db := database.GetDB()
	if err := db.AutoMigrate(&model.User{}, &model.Token{}, &model.RefreshToken{}, &model.RecallRecord{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	authService := service.NewAuthService(cfg)
	recallService := service.NewRecallService()
	notifyService := service.NewNotifyService(cfg)

	authHandler := handler.NewAuthHandler(authService)
	recallHandler := handler.NewRecallHandler(recallService, notifyService, authService)
	notifyHandler := handler.NewNotifyHandler(notifyService)

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Static("/web", "./web")
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/web/index.html")
	})

	router.GET("/recall", recallHandler.HandleRecall)

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/create-jwt-token", authHandler.CreateJWTToken)         // 用ApiToken换取JWT
			auth.POST("/refresh-jwt-by-refresh-token", authHandler.RefreshJWTByRefreshToken) // 用RefreshToken刷新JWT
			auth.POST("/refresh-jwt-by-api-token", authHandler.RefreshJWTByApiToken)         // 用ApiToken刷新JWT
			auth.POST("/logout", middleware.AuthMiddleware(authService), authHandler.Logout)
			auth.POST("/refresh", middleware.AuthMiddleware(authService), authHandler.Refresh)
			auth.POST("/send-code", authHandler.SendRegisterCode)
			auth.POST("/send-reset-code", authHandler.SendResetCode)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.GET("/query", recallHandler.Query)
			protected.GET("/query/latest", recallHandler.QueryLatest)
			protected.GET("/history", recallHandler.QueryHistory)

			notify := protected.Group("/notify")
			{
				notify.POST("/set", notifyHandler.SetNotifyURL)
				notify.GET("/get", notifyHandler.GetNotifyURL)
			}

			account := protected.Group("/account")
			{
				account.POST("/change-password", authHandler.ChangePassword)
				account.POST("/delete", authHandler.DeleteAccount)
				account.GET("/get-api-token", authHandler.GetApiToken)       // 获取ApiToken
				account.POST("/update-api-token", authHandler.UpdateApiToken) // 更换ApiToken
			}

			// JWT Token管理
			token := protected.Group("/token")
			{
				token.GET("/info", authHandler.GetJWTTokenInfo) // 获取JWT和RefreshToken信息
			}
		}
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	tokenRepo := repository.NewTokenRepository()
	refreshTokenRepo := repository.NewRefreshTokenRepository()
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		// 启动时先清理一次
		if err := tokenRepo.DeleteExpiredTokens(); err != nil {
			log.Printf("Failed to cleanup expired tokens: %v", err)
		}
		if err := refreshTokenRepo.DeleteExpiredTokens(); err != nil {
			log.Printf("Failed to cleanup expired refresh tokens: %v", err)
		}
		for range ticker.C {
			if err := tokenRepo.DeleteExpiredTokens(); err != nil {
				log.Printf("Failed to cleanup expired tokens: %v", err)
			}
			if err := refreshTokenRepo.DeleteExpiredTokens(); err != nil {
				log.Printf("Failed to cleanup expired refresh tokens: %v", err)
			}
		}
	}()

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
