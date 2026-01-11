package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	pb "github.com/JanMird/gateway/gen/auth"
	"github.com/JanMird/gateway/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcAuthClient struct {
	client  pb.AuthServiceClient
	timeout time.Duration
}

func NewGRPCAuthClient(client GRPCClient, timeout time.Duration) *grpcAuthClient {
	return &grpcAuthClient{
		client:  pb.NewAuthServiceClient(client),
		timeout: timeout,
	}
}

func (c *grpcAuthClient) Authenticate(ctx context.Context, u domain.UserUnsafe) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &pb.AuthenticateRequest{
		Password: u.Password,
	}

	if u.Email != "" {
		req.Email = &u.Email
	}
	if u.Phone != "" {
		req.Phone = &u.Phone
	}

	resp, err := c.client.Authenticate(ctx, req)
	if err != nil {
		if status, ok := status.FromError(err); ok {
			switch status.Code() {
			case codes.DeadlineExceeded:
				return "", "", fmt.Errorf("authentication timeout: %v", err)
			default:
				return "", "", fmt.Errorf("gRPC error (%s): %v", status.Code().String(), status.Message())
			}
		}
		return "", "", fmt.Errorf("authentication failed: %v", err)
	}

	if resp == nil {
		return "", "", errors.New("empty response from auth service")
	}

	if resp.AccessToken == "" {
		return "", "", errors.New("access token not found in response")
	}

	if resp.RefreshToken == "" {
		return "", "", errors.New("refresh token not found in response")
	}

	return resp.AccessToken, resp.RefreshToken, nil
}

func (c *grpcAuthClient) SignUp(ctx context.Context, u domain.UserUnsafe) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &pb.SignUpRequest{
		Password: u.Password,
	}

	if u.Email != "" {
		email := u.Email
		req.Email = &email
	}

	if u.Phone != "" {
		phone := u.Phone
		req.Phone = &phone
	}

	resp, err := c.client.SignUp(ctx, req)
	if err != nil {
		if status, ok := status.FromError(err); ok {
			switch status.Code() {
			case codes.DeadlineExceeded:
				return "", "", fmt.Errorf("signup timeout: %w", err)
			case codes.AlreadyExists:
				return "", "", fmt.Errorf("user already exists: %w", err)
			case codes.DataLoss:
				return "", "", fmt.Errorf("user created, token not created: %w", status.Err())
			default:
				return "", "", fmt.Errorf("signup failed (%s): %v", status.Code().String(), status.Message())
			}
		}
		return "", "", fmt.Errorf("signup failed: %v", err)
	}

	if resp == nil {
		return "", "", errors.New("empty response from signup service")
	}

	if resp.AccessToken == "" {
		return "", "", errors.New("access token not found in response")
	}

	if resp.RefreshToken == "" {
		return "", "", errors.New("refresh token not found in response")
	}

	return resp.AccessToken, resp.RefreshToken, nil
}

func (c *grpcAuthClient) RefreshToken(ctx context.Context, token string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if token == "" {
		return "", "", errors.New("refresh token cannot be empty")
	}

	req := &pb.RefreshTokenRequest{
		RefreshToken: token,
	}

	resp, err := c.client.RefreshToken(ctx, req)
	if err != nil {
		if status, ok := status.FromError(err); ok {
			switch status.Code() {
			case codes.DeadlineExceeded:
				return "", "", fmt.Errorf("refresh token timeout: %v", err)
			default:
				return "", "", fmt.Errorf("refresh failed (%s): %v", status.Code().String(), status.Message())
			}
		}
		return "", "", fmt.Errorf("refresh token failed: %v", err)
	}

	if resp == nil {
		return "", "", errors.New("empty response from refresh token service")
	}

	if resp.AccessToken == "" {
		return "", "", errors.New("access token not found in refresh response")
	}

	return resp.AccessToken, resp.RefreshToken, nil
}

func (c *grpcAuthClient) GetClaimsFromToken(ctx context.Context, token string) (domain.Claims, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if token == "" {
		return domain.Claims{}, errors.New("access token cannot be empty")
	}

	req := &pb.GetClaimsFromTokenRequest{
		AccessToken: token,
	}

	resp, err := c.client.GetClaimsFromToken(ctx, req)
	if err != nil {
		if status, ok := status.FromError(err); ok {
			switch status.Code() {
			case codes.DeadlineExceeded:
				return domain.Claims{}, fmt.Errorf("token validation timeout: %v", err)
			default:
				return domain.Claims{}, fmt.Errorf("token validation failed (%s): %v", status.Code().String(), status.Message())
			}
		}
		return domain.Claims{}, fmt.Errorf("token validation failed: %v", err)
	}

	if resp == nil {
		return domain.Claims{}, errors.New("empty response from token validation service")
	}

	if resp.Id == "" {
		return domain.Claims{}, errors.New("user ID not found in claims response")
	}

	userID, err := strconv.Atoi(resp.Id)
	if err != nil {
		return domain.Claims{}, fmt.Errorf("invalid user ID format in claims: %v", err)
	}

	roleInt := int(resp.Role)

	claims := domain.Claims{
		Sub:  userID,
		Role: roleInt,
		Exp:  0,
		Iat:  time.Time{},
	}

	return claims, nil
}

func (c *grpcAuthClient) SignOut(ctx context.Context, token string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if token == "" {
		return errors.New("token cannot be empty for sign out")
	}

	req := &pb.SignOutRequest{
		Token: token,
	}

	_, err := c.client.SignOut(ctx, req)
	if err != nil {
		if status, ok := status.FromError(err); ok {
			switch status.Code() {
			case codes.DeadlineExceeded:
				return fmt.Errorf("sign out timeout: %v", err)
			default:
				return fmt.Errorf("sign out failed (%s): %v", status.Code().String(), status.Message())
			}
		}
		return fmt.Errorf("sign out failed: %v", err)
	}

	return nil
}
