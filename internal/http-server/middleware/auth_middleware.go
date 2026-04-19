package middleware

import (
	"internship/internal/domain"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	CtxUserID = "userID"
	CtxRole   = "role"
)

type TokenParser interface {
	ParseToken(tokenString string) (userID, role string, err error)
}

func AuthMiddleware(parser TokenParser, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")

		if !strings.HasPrefix(header, "Bearer ") {
			logger.Debug("auth middleware: missing bearer token",
				zap.String("path", c.FullPath()),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "missing or invalid authorization header",
				},
			})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")

		userID, role, err := parser.ParseToken(tokenStr)
		if err != nil {
			logger.Debug("auth middleware: invalid token",
				zap.String("path", c.FullPath()),
				zap.Error(err),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "invalid or expired token",
				},
			})
			return
		}

		c.Set(CtxUserID, userID)
		c.Set(CtxRole, role)

		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {

	return func(c *gin.Context) {
		role_in, exists := c.Get(CtxRole)
		if !exists || role_in != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    domain.ErrForbidden,
					"message": "FORBIDDEN",
				},
			})
			return
		}

		c.Next()
	}
}
