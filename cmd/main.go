package main

import (
	"os"
	"strconv"
	"time"

	"github.com/JanMird/gateway/internal/client"
	"github.com/JanMird/gateway/internal/domain"
	"github.com/JanMird/gateway/internal/handler"
	"github.com/JanMird/gateway/internal/middleware"
	"github.com/JanMird/gateway/internal/service"
	"github.com/JanMird/gateway/pkg/cache"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	_ = godotenv.Load(".env")

	var log = logrus.New()
	log.SetLevel(logrus.DebugLevel)

	redisEndpoint := os.Getenv("REDIS_ENDPOINT")
	redisUser := os.Getenv("RESID_USER")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		panic("invalid redis_db")
	}
	redisMaxRetries, err := strconv.Atoi(os.Getenv("REDIS_MAX_RETRIES"))
	if err != nil {
		panic("invalid redis_max_retries")
	}
	redisDialTimeout, err := strconv.Atoi(os.Getenv("REDIS_DIAL_TIMEOUT"))
	if err != nil {
		panic("invalid redis_dial_timeout")
	}
	redisTimeout, err := strconv.Atoi(os.Getenv("REDIS_TIMEOUT"))
	if err != nil {
		panic("invalid redis_timeout")
	}

	cfg := cache.Config{
		Addr:        redisEndpoint,
		Password:    redisPassword,
		User:        redisUser,
		DB:          redisDB,
		MaxRetries:  redisMaxRetries,
		DialTimeout: time.Duration(redisDialTimeout) * time.Second,
		Timeout:     time.Duration(redisTimeout) * time.Second,
	}

	redisClient, err := cache.NewRedisClient(cfg)
	if err != nil {
		panic("redis client creation error")
	}
	cacheClient := cache.New(redisClient)

	connCRM, err := grpc.NewClient(os.Getenv("CRM_URL"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer connCRM.Close()
	if err != nil {
		panic("Failed to connect to users service: %v")
	}
	connCRM.Connect()

	bookingClient := client.NewGRPCBookingClient(connCRM, domain.GRPCTimeout)
	bookingSrvice := service.NewBookingService(bookingClient)
	bookingHandler := handler.NewBookingHandler(bookingSrvice, log)

	connAuth, err := grpc.NewClient(os.Getenv("AUTH_URL"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer connAuth.Close()
	if err != nil {
		panic("Failed to connect to users service: %v")
	}
	connAuth.Connect()

	authClient := client.NewGRPCAuthClient(connAuth, domain.GRPCTimeout)
	authService := service.NewAuthService(authClient, service.WithCache(cacheClient), service.WithLogger(log))
	authHandler := handler.NewAuthHandler(authService, log)

	r := gin.Default()

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/authenticate", authHandler.Authenticate)
		authGroup.POST("/signup", authHandler.SignUp)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.GET("/health", authHandler.HealthCheck)
		authGroup.POST("/signout", authHandler.SignOut)
	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		BookingGroup := protected.Group("/book")
		{
			BookingGroup.POST("/booking", bookingHandler.CreateBooking)
		}
	}

	r.Run(":8081")
}
