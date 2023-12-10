package middleware

import (
	"context"
	"github.com/gin-gonic/gin"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/keycloak"
)

func TokenSource() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenSource := keycloak.AuthHeaderTokenSource(c.Request.Header.Get("Authorization"))
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), common.CtxTokenSource, tokenSource))

		c.Next()
	}
}
