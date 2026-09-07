package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	adminSecret string
}

func NewAuthMiddleware(adminSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		adminSecret: adminSecret,
	}
}

func (m *AuthMiddleware) AdminSecretMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminToken := c.GetHeader("X-Admin-Secret")

		if adminToken != m.adminSecret {
			abortWithError(c, http.StatusForbidden, "access denied", "AUTHZ_ROLE_FORBIDDEN")
			return
		}

		c.Next()
	}
}

func abortWithError(c *gin.Context, status int, message, code string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": message,
		"code":  code,
	})
}
