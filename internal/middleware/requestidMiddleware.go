package middleware

import (
	"context"

	"go-rest-example/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReqIDMiddleware injects a request ID into the context and response header, creates one if it is not present already.
func ReqIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.Request.Header.Get(util.RequestIdentifier)
		if requestId == "" {
			requestId = uuid.New().String()
		}

		ctx := context.WithValue(c.Request.Context(), util.RequestIdentifier, requestId)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set(util.RequestIdentifier, requestId)
		c.Next()
	}
}