package handler

import (
	"context"
	"net/http"

	"github.com/JanMird/gateway/internal/domain"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Authenticate(ctx context.Context, u domain.UserUnsafe) (string, string, error)
	SignUp(ctx context.Context, u domain.UserUnsafe) (string, string, error)
	RefreshToken(ctx context.Context, token string) (string, string, error)
	GetClaimsFromToken(ctx context.Context, token string) (domain.Claims, error)
	SignOut(ctx context.Context, token string) error
}

type authHandler struct {
	service AuthService
	logger  Logger
}

func NewAuthHandler(service AuthService, logger Logger) *authHandler {
	return &authHandler{
		service: service,
		logger:  logger,
	}
}

func (h *authHandler) Authenticate(ctx *gin.Context) {
	ctxx := ctx.Request.Context()
	var req domain.UserUnsafe

	if err := ctx.BindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Email == "" && req.Phone == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Email or phone is required"})
		return
	}

	if req.Password == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	accessToken, refreshToken, err := h.service.Authenticate(ctxx, req)
	if err != nil {
		h.logger.Error("Authentication failed:", err)

		switch {
		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *authHandler) SignUp(ctx *gin.Context) {
	ctxx := ctx.Request.Context()
	var req domain.UserUnsafe

	if err := ctx.BindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Email == "" && req.Phone == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Email or phone is required"})
		return
	}

	if req.Password == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	accessToken, refreshToken, err := h.service.SignUp(ctxx, req)
	if err != nil {
		h.logger.Error("SignUp failed:", err)
		st, ok := status.FromError(err)
		switch {
		case ok && st.Code() == codes.DataLoss:
			ctx.AbortWithStatusJSON(http.StatusCreated, gin.H{"error": err.Error()})
		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"message":       "User registered successfully",
	})
}

func (h *authHandler) RefreshToken(ctx *gin.Context) {
	ctxx := ctx.Request.Context()

	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	accessToken, refreshToken, err := h.service.RefreshToken(ctxx, req.RefreshToken)
	if err != nil {
		h.logger.Error("Refresh token failed:", err)

		switch {
		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Token refresh failed"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *authHandler) SignOut(ctx *gin.Context) {
	ctxx := ctx.Request.Context()

	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization format. Expected 'Bearer <token>'"})
		return
	}

	token := authHeader[7:]

	err := h.service.SignOut(ctxx, token)
	if err != nil {
		h.logger.Error("SignOut failed:", err)

		switch {
		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Sign out failed"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully signed out"})
}

func (h *authHandler) GetUserInfo(ctx *gin.Context) {
	ctxx := ctx.Request.Context()

	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
		return
	}

	token := authHeader[7:]

	claims, err := h.service.GetClaimsFromToken(ctxx, token)
	if err != nil {
		h.logger.Error("Get user info failed:", err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user_id": claims.Sub,
		"role":    claims.Role,
		"expires": claims.Exp,
	})
}

// HealthCheck is a simple endpoint to verify the auth service is working
func (h *authHandler) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth",
	})
}
