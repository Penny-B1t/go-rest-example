package middleware

import (
	"errors"
	error2 "go-rest-example/internal/error" // 우리가 만든 error 패키지
	"go-rest-example/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler(lgr *logger.AppLogger) gin.HandlerFunc {
	return func(c *gin.Context){
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors[0].Err

		var ae *error2.AppError

		if errors.As(err, &ae){
			lgr.Error().Err(ae.Err).Int("status", ae.StatusCode).Msg(ae.Message)
			c.AbortWithStatusJSON(ae.StatusCode, gin.H{
				"message":   ae.Message,
			})
			return
		}

		lgr.Error().Err(err).Msg("an unexpected error occurred")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "서버 내부 오류가 발생했습니다.",
		})
	}
}