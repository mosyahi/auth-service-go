package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/pkg/database"
)

func main() {
	cfg := config.Load()

	// 1. Infrastructure Initialization
	db, err := database.Connect(cfg.DBURL)
	if err != nil {
		log.Fatalf("❌ Database failure: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	// Run migrations
	if err := db.AutoMigrate(&domain.User{}, &domain.UserSession{}); err != nil {
		log.Fatalf("❌ Migration failure: %v", err)
	}

	// 2. Dependency Injection (Wiring)
	userRepo := repository.NewUserRepository(db)
	// Pastikan NewAuthService menerima rdb sekarang
	authService := service.NewAuthService(userRepo, rdb, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	// 3. Router Setup
	router := setupRouter(cfg, authHandler, rdb)

	// 4. Server Start & Graceful Shutdown
	runServer(router)
}

func setupRouter(cfg *config.Config, authHandler *handler.AuthHandler, rdb *redis.Client) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	v1 := r.Group("/api/v1")
	{
		// Authentication & Public Routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", middleware.RateLimiter(rdb), authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// User Protected Routes
		user := v1.Group("/user")
		user.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			user.GET("/me", authHandler.GetProfile)

			// Admin Section (RBAC)
			admin := user.Group("/admin")
			admin.Use(middleware.AuthorizeRole("admin"))
			{
				admin.GET("/users", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Admin Access Granted"})
				})
			}
		}
	}
	return r
}

func runServer(router *gin.Engine) {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("🚀 Auth Service active on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful Shutdown Logic
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ Server forced to shutdown:", err)
	}
	log.Println("👋 Server exited gracefully")
}
