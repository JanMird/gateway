package main

import (
	"os"

	"github.com/JanMird/gateway/internal/client"
	"github.com/JanMird/gateway/internal/domain"
	"github.com/JanMird/gateway/internal/handler"
	"github.com/JanMird/gateway/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	_ = godotenv.Load(".env")

	var log = logrus.New()

	connUsers, err := grpc.NewClient(os.Getenv("USERS_SERVICE_URL"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer connUsers.Close()
	if err != nil {
		panic("Failed to connect to users service: %v")
	}
	connUsers.Connect()

	client := client.NewGRPCBookingClient(connUsers, domain.GRPCTimeout)

	service := service.NewBookingService(client)

	handler := handler.NewBookingHandler(service, log)

	r := gin.Default()

	r.POST("/booking", handler.CreateBooking)

	r.Run(":8081")
}
