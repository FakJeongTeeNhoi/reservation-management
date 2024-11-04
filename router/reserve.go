package router

import (
	"github.com/FakJeongTeeNhoi/reservation-management/controller"
	"github.com/gin-gonic/gin"
)

func ReserveRouterGroup(server *gin.RouterGroup) {
	reserve := server.Group("/reserve")

	reserve.POST("/", controller.CreateReservationHandler)
	reserve.GET("/", controller.GetReservationsHandler)
	reserve.GET("/:reservationId", controller.GetReservationHandler)
	reserve.PUT("/:reservationId", controller.CancelReservationHandler)
	reserve.PUT("/confirm", controller.ConfirmReservationHandler)
}
