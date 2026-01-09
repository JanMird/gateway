package handler

import (
	"context"
	"net/http"

	"github.com/JanMird/gateway/internal/domain"
	"github.com/gin-gonic/gin"
)

type Logger interface {
	Debug(...interface{})
	Error(...interface{})
	Info(...interface{})
}

type BookingService interface {
	CreateBooking(context.Context, *domain.Booking) (int, error)
}

type handler struct {
	service BookingService
	logger  Logger
}

func NewBookingHandler(service BookingService, logger Logger) *handler {
	return &handler{
		service: service,
		logger:  logger,
	}
}

func (h *handler) CreateBooking(ctx *gin.Context) {
	ctxx := ctx.Request.Context()
	var req domain.Booking

	if err := ctx.BindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateBooking(ctxx, &req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
