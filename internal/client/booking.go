package client

import (
	"context"
	"fmt"
	"time"

	"github.com/JanMird/gateway/internal/domain"

	"github.com/JanMird/gateway/gen/booking"
	pb "github.com/JanMird/gateway/gen/booking"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCClient interface {
	grpc.ClientConnInterface
}

type grpcBookingClient struct {
	client  pb.BookingServiceClient
	timeout time.Duration
}

func NewGRPCBookingClient(client GRPCClient, timeout time.Duration) *grpcBookingClient {
	return &grpcBookingClient{
		client:  pb.NewBookingServiceClient(client),
		timeout: timeout,
	}
}

func (c *grpcBookingClient) CreateBooking(ctx context.Context, req *domain.Booking) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	protoReq := &pb.CreateBookingRequest{
		AdminId:     req.AdminID,
		ClientId:    req.ClientID,
		BookingType: req.BookingType,
		StartTime:   timestamppb.New(req.StartTime),
		EndTime:     timestamppb.New(req.EndTime),
		TrainerId:   req.TrainerID,
		LessonId:    req.LessonID,
		Status:      req.Status,
		Amount:      req.Amount,
		Currency:    req.Currency,
	}

	protoReq.EquipmentItems = make([]*booking.EquipmentItem, len(req.EquipmentItems))
	for i, item := range req.EquipmentItems {
		protoReq.EquipmentItems[i] = &booking.EquipmentItem{
			EquipmentId: item.EquipmentID,
			UnitId:      item.UnitID,
		}
	}

	protoResp, err := c.client.CreateBooking(ctx, protoReq)
	if err != nil {
		return -1, fmt.Errorf("gRPC call failed: %w", err)
	}

	return int(protoResp.BookingId), nil
}
