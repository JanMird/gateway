package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/JanMird/gateway/internal/domain"
)

type AuthService interface {
	Authenticate(context.Context, domain.UserUnsafe) (string, string, error)
	SignUp(context.Context, domain.UserUnsafe) (string, string, error)
	RefreshToken(context.Context, string) (string, string, error)
	GetClaimsFromToken(context.Context, string) (domain.Claims, error)
	SignOut(context.Context, string) error
}

func AuthMiddleware(authService AuthService, requiredRole ...int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		token := parts[1]

		claims, err := authService.GetClaimsFromToken(ctx.Request.Context(), token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		if len(requiredRole) > 0 {
			hasRequiredRole := false
			for _, role := range requiredRole {
				if claims.Role == role {
					hasRequiredRole = true
					break
				}
			}

			if !hasRequiredRole {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
				return
			}
		}

		ctx.Set("user_id", claims.Sub)
		ctx.Set("user_role", claims.Role)

		ctx.Next()
	}
}
