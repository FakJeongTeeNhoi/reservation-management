package middleware

import (
	"github.com/FakJeongTeeNhoi/reservation-management/model/response"
	"github.com/gin-gonic/gin"
)

func AuthorizedStaff() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetHeader("user_id")

		if userId != "" {
			response.Unauthorized("Unauthorized").AbortWithError(c)
			return
		}

		c.Next()
	}
}
