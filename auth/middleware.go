package auth

import (
	"net/http"
	"strings"

	"cosine/config"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token and sets user claims in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GlobalConfig
		if cfg == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "config not loaded"})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := ParseToken(tokenString, cfg.JWT.Secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Set claims in context for use in handlers
		c.Set("claims", claims)
		c.Set("linuxdo_id", claims.LinuxDoID)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// GetClaimsFromContext retrieves JWT claims from gin context
func GetClaimsFromContext(c *gin.Context) (*JWTClaims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}
	jwtClaims, ok := claims.(*JWTClaims)
	return jwtClaims, ok
}
