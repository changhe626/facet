package middleware

import (
	"facet-backend/internal/api/response"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the JWT claims (must be consistent with the token package definition)
type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTMiddleware creates a Gin middleware for JWT validation
func JWTMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.CustomError(c, http.StatusUnauthorized, "authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.CustomError(c, http.StatusUnauthorized, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			response.CustomError(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		if !token.Valid {
			response.CustomError(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		// Put user info into Gin's context (using constants from context_keys.go)
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(TenantIDKey), claims.TenantID)
		c.Set(string(UserRoleKey), claims.Role)

		c.Next()
	}
}
