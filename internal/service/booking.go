package service

import (
	"context"
	"fmt"

	"github.com/JanMird/gateway/internal/domain"
)

type BookingClient interface {
	CreateBooking(ctx context.Context, req *domain.Booking) (int, error)
}

type bookingService struct {
	client BookingClient
}

func NewBookingService(client BookingClient) *bookingService {
	return &bookingService{
		client: client,
	}
}

func (s *bookingService) CreateBooking(ctx context.Context, req *domain.Booking) (int, error) {
	res, err := s.client.CreateBooking(ctx, req)
	if err != nil {
		return -1, fmt.Errorf("CreateBookingInternal: %w", err)
	}

	return res, nil
}
