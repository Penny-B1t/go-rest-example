package middleware

import "github.com/gin-gonic/gin"

// 인증 헤더 유효성 검사 필요
// 보안을 위한 암호화된 값 규격 필요
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context){
		c.Next()
	}
}