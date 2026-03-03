package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JanMird/gateway/internal/domain"
)

type AuthClient interface {
	Authenticate(ctx context.Context, u domain.UserUnsafe) (string, string, error)
	SignUp(ctx context.Context, u domain.UserUnsafe) (string, string, error)
	RefreshToken(ctx context.Context, token string) (string, string, error)
	GetClaimsFromToken(ctx context.Context, token string) (domain.Claims, error)
	SignOut(ctx context.Context, token string) error
}

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
}

type authService struct {
	client AuthClient
	cache  Cache
	logger Logger
}

type Options struct {
	Cache         Cache
	Logger        Logger
	TokenCacheTTL time.Duration
}

func NewAuthService(client AuthClient, opts ...func(*authService)) *authService {
	s := &authService{
		client: client,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithCache(cache Cache) func(*authService) {
	return func(s *authService) {
		s.cache = cache
	}
}

func WithLogger(logger Logger) func(*authService) {
	return func(s *authService) {
		s.logger = logger
	}
}

func (s *authService) Authenticate(ctx context.Context, u domain.UserUnsafe) (string, string, error) {
	if err := validateCredentials(u); err != nil {
		return "", "", err
	}

	accessToken, refreshToken, err := s.client.Authenticate(ctx, u)
	if err != nil {
		s.logError("authentication failed", err)

		return "", "", fmt.Errorf("authentication failed: %w", err)
	}

	s.logInfo("user authenticated", "email", u.Email, "phone", u.Phone)
	return accessToken, refreshToken, nil
}

func (s *authService) SignUp(ctx context.Context, u domain.UserUnsafe) (string, string, error) {
	if err := validateSignUp(u); err != nil {
		return "", "", err
	}

	accessToken, refreshToken, err := s.client.SignUp(ctx, u)
	if err != nil {
		s.logError("signup failed", err)
		return "", "", fmt.Errorf("sign up failed: %w", err)
	}

	s.logDebug("user registered", "email", u.Email, "phone", u.Phone, "role", u.Role)
	return accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(ctx context.Context, token string) (string, string, error) {
	if token == "" {
		return "", "", errors.New("refresh token is required")
	}

	accessToken, refreshToken, err := s.client.RefreshToken(ctx, token)
	if err != nil {
		s.logError("token refresh failed", err)

		return "", "", fmt.Errorf("token refresh failed: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *authService) GetClaimsFromToken(ctx context.Context, token string) (domain.Claims, error) {
	if token == "" {
		return domain.Claims{}, errors.New("token is required")
	}

	if s.cache != nil {
		cacheKey := fmt.Sprintf("token:%s", token)
		if cachedClaims, err := s.cache.Get(ctx, cacheKey); err == nil && cachedClaims != nil {
			tmp := strings.Split(string(cachedClaims), " ")
			if len(tmp) != 2 {
				return domain.Claims{}, errors.New("invalid cached data")
			}

			id, err := strconv.Atoi(tmp[0])
			if err != nil {
				return domain.Claims{}, errors.New("invalid token")
			}
			role, err := strconv.Atoi(tmp[1])
			if err != nil {
				return domain.Claims{}, errors.New("invalid role")
			}

			s.logDebug("cache hit for token validation")
			return domain.Claims{
				Sub:  id,
				Role: role,
			}, nil
		}
	}

	claims, err := s.client.GetClaimsFromToken(ctx, token)
	if err != nil {
		s.logError("token validation failed", err)

		return domain.Claims{}, fmt.Errorf("token validation failed: %w", err)
	}

	if s.cache != nil {
		cacheKey := fmt.Sprintf("token:%s", token)
		cacheVal := fmt.Sprintf("%d %d", claims.Sub, claims.Role)
		s.cache.Set(ctx, cacheKey, []byte(cacheVal), 5*time.Minute)
	}

	return claims, nil
}

func (s *authService) SignOut(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is required")
	}

	err := s.client.SignOut(ctx, token)
	if err != nil {
		s.logError("signout failed", err)
		return fmt.Errorf("signout failed: %w", err)
	}

	s.logInfo("user signed out successfully")
	return nil
}

func validateCredentials(u domain.UserUnsafe) error {
	if u.Password == "" {
		return errors.New("password is required")
	}

	if u.Email == "" && u.Phone == "" {
		return errors.New("email or phone is required")
	}

	return nil
}

func validateSignUp(u domain.UserUnsafe) error {
	if err := validateCredentials(u); err != nil {
		return err
	}

	if u.Email != "" && !isValidEmail(u.Email) {
		return errors.New("invalid email format")
	}

	if u.Phone != "" && !isValidPhone(u.Phone) {
		return errors.New("invalid phone format")
	}

	if len(u.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	return nil
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func isValidPhone(phone string) bool {
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	return len(cleaned) >= 10
}

func (s *authService) logError(msg string, err error) {
	if s.logger != nil {
		s.logger.Error(msg, "error", err)
	}
}

func (s *authService) logInfo(msg string, fields ...interface{}) {
	if s.logger != nil {
		s.logger.Info(append([]interface{}{msg}, fields...)...)
	}
}

func (s *authService) logDebug(msg string, fields ...interface{}) {
	if s.logger != nil {
		s.logger.Debug(append([]interface{}{msg}, fields...)...)
	}
}
