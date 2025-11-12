package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/internal/circuit"
	"api-gateway/internal/config"
	"api-gateway/internal/handler"
	"api-gateway/internal/middleware"
	"api-gateway/internal/service"
	"api-gateway/pkg/logger"
	"api-gateway/pkg/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.NewLogger(cfg.Logging.Level)
	defer log.Sync()

	log.Info("Starting API Gateway")

	mongoClient, err := storage.NewMongoClient(cfg.MongoDB)
	if err != nil {
		log.Fatal("MongoDB connection failed", "error", err)
	}
	defer mongoClient.Close()

	redisClient, err := storage.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatal("Redis connection failed", "error", err)
	}
	defer redisClient.Close()

	log.Info("Database connections established")

	registry := service.NewRegistry(cfg.Services)
	loadBalancer := service.NewLoadBalancer()
	breakerManager := circuit.NewBreakerManager(cfg.CircuitBreaker)

	authHandler := handler.NewAuthHandler(mongoClient, cfg, log)
	proxyHandler := handler.NewProxyHandler(registry, loadBalancer, breakerManager, log)
	healthHandler := handler.NewHealthHandler(redisClient, mongoClient)

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.Recovery(log))
	router.Use(middleware.RequestLogger(log))
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS(cfg.CORS))
	router.Use(middleware.SecurityHeaders())

	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Readiness)

	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	api := router.Group("/api/v1")
	api.Use(middleware.RateLimiter(redisClient, cfg.RateLimit))
	api.Use(middleware.JWTAuth(cfg.JWT.Secret))
	{
		api.GET("/profile", authHandler.GetProfile)
		
		// Specific proxy routes for each service (instead of wildcard)
		api.Any("/users/*path", proxyHandler.ProxyRequest)
		api.Any("/products/*path", proxyHandler.ProxyRequest)
		api.Any("/orders/*path", proxyHandler.ProxyRequest)
	}

	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.RateLimiter(redisClient, cfg.RateLimit))
	admin.Use(middleware.JWTAuth(cfg.JWT.Secret))
	admin.Use(middleware.RoleAuth("admin"))
	{
		admin.GET("/services", proxyHandler.ListServices)
		admin.POST("/services", proxyHandler.RegisterService)
		admin.DELETE("/services/:name", proxyHandler.UnregisterService)
	}

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Timeouts.Read) * time.Second,
		WriteTimeout:   time.Duration(cfg.Timeouts.Write) * time.Second,
		IdleTimeout:    time.Duration(cfg.Timeouts.Idle) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Info("Server started", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Shutdown failed", "error", err)
	}

	log.Info("Gateway stopped")
}
