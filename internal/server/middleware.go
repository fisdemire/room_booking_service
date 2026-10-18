package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/gin-gonic/gin"
)

type TokenParser interface {
	Parse(
		ctx context.Context,
		token string,
	) (string, domain.Role, error)
}

const (
	userIDKey   = "auth_user_id"
	userRoleKey = "auth_user_role"
)

func AuthMiddleware(parser TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.Fields(c.GetHeader("Authorization"))

		if len(header) != 2 ||
			!strings.EqualFold(header[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or invalid authorization header",
			})
			return
		}

		userID, role, err := parser.Parse(
			c.Request.Context(),
			header[1],
		)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		c.Set(userIDKey, userID)
		c.Set(userRoleKey, role)

		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get(userRoleKey)
		if !ok || role != domain.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			return
		}

		c.Next()
	}
}

func UserIDFromContext(c *gin.Context) (string, bool) {
	value, ok := c.Get(userIDKey)
	if !ok {
		return "", false
	}

	userID, ok := value.(string)
	return userID, ok && userID != ""
}

func RoleFromContext(c *gin.Context) (domain.Role, bool) {
	value, ok := c.Get(userRoleKey)
	if !ok {
		return "", false
	}

	role, ok := value.(domain.Role)
	return role, ok
}
