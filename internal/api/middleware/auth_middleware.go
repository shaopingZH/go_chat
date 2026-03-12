package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go-chat/pkg/jwtutil"
)

const ContextUserIDKey = "user_id"

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			c.Abort()
			return
		}

		//验证token
		claims, err := jwtutil.ParseToken(secret, strings.TrimSpace(parts[1]))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, false
	}

	userID, ok := value.(uint64)
	if !ok {
		return 0, false
	}

	return userID, true
}
