package controller

import (
	"github.com/FakJeongTeeNhoi/reservation-management/model"
	"github.com/FakJeongTeeNhoi/reservation-management/model/response"
	"github.com/FakJeongTeeNhoi/reservation-management/service"
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
)

func checkAvailability(roomId uint, start string, end string) bool {
	reservations := model.Reservations{}
	if err := reservations.GetByRoomIdAndTimePeriod(roomId, start, end); err != nil {
		return false
	}

	return len(reservations) == 0
}

func CreateReservationHandler(c *gin.Context) {
	userType := c.GetHeader("type")
	if userType != "staff" && userType != "user" {
		response.BadRequest("Invalid user type").AbortWithError(c)
		return
	}

	rcr := model.ReservationCreateRequest{}
	if err := c.ShouldBindJSON(&rcr); err != nil {
		response.BadRequest("Invalid request").AbortWithError(c)
		return
	}

	if !checkAvailability(rcr.RoomId, rcr.StartDateTime, rcr.EndDateTime) {
		response.BadRequest("Room is not available").AbortWithError(c)
		return
	}

	if !service.IsCurrentTimeBefore(rcr.StartDateTime, -30){
		response.BadRequest("You can only make reservation at least 30 minutes before the start time").AbortWithError(c)
		return
	}

	reservation := rcr.ToReservation()

	reservation.Status = "created"

	if(userType == "staff") {
		reservation.Participants = reservation.PendingParticipants
		reservation.PendingParticipants = []int64{}
	} else {
		for i, pendingParticipant := range reservation.PendingParticipants {
			if pendingParticipant == service.ParseToInt64(c.GetHeader("user_id")) {
				reservation.Participants = append(reservation.Participants, pendingParticipant)
				reservation.PendingParticipants = append(reservation.PendingParticipants[:i], reservation.PendingParticipants[i+1:]...)
				break
			}
		}
	}

	if err := reservation.Create(); err != nil {
		response.InternalServerError("Failed to create reservation").AbortWithError(c)
		return
	}

	if(userType == "staff") {
		c.JSON(200, response.CommonResponse{
			Success: true,
		}.AddInterfaces(map[string]interface{}{
			"reservation": reservation,
		}))
		return
	}

	// send email to every pending participant
	subject := "You have been invited to a reservation"
	for _, pendingParticipant := range reservation.PendingParticipants {
		receiverEmail := strconv.FormatInt(pendingParticipant, 10) + os.Getenv("EMAIL_DOMAIN")

		body := "You have been invited to a reservation. Please click the link below to view the reservation details.\n" +
			"<br> Please validate your account by clicking the link below: <a href='" +
			os.Getenv("FRONTEND_URL") +
			os.Getenv("RESERVATION_VERIFY_PATH") +
			"?reservationId=" +
			strconv.FormatUint(uint64(reservation.ID), 10) +
			"&userId=" +
			strconv.FormatUint(uint64(pendingParticipant), 10) +
			"'>View Reservation</a>"

		_ = service.SendMail(receiverEmail, subject, body)
	}

	c.JSON(200, response.CommonResponse{
		Success: true,
	}.AddInterfaces(map[string]interface{}{
		"reservation": reservation,
	}))


	// TODO: Flow data to message broker
	// call grpc to get room info
	// send data via message broker
}

func GetReservationsHandler(c *gin.Context) {
	if c.GetHeader("user_id") != "" && c.GetHeader("user_id") != c.Query("userId") {
		response.Forbidden("You are not allowed to view other user's reservations").AbortWithError(c)
		return
	}

	userId := c.Query("userId")
	if userId != "" {
		userId = c.GetHeader("user_id")
	}

	roomId := c.Query("roomId")

	filters := map[string]interface{}{}
	if roomId != "" {
		filters["room_id"] = roomId
	}

	unfilteredReservations := model.Reservations{}

	if err := unfilteredReservations.GetAll(nil); err != nil {
		response.InternalServerError("Failed to get reservations").AbortWithError(c)
		return
	}

	reservations := model.Reservations{}

	if userId != "" {
		// filter only reservations that user is in participants or pending participants
		for _, reservation := range unfilteredReservations {
			for _, participant := range reservation.Participants {
				if participant == service.ParseToInt64(userId) {
					reservations = append(reservations, reservation)
					break
				}
			}
			for _, pendingParticipant := range reservation.PendingParticipants {
				if pendingParticipant == service.ParseToInt64(userId) {
					reservations = append(reservations, reservation)
					break
				}
			}
		}
	}

	// TODO: call grpc to get room info for each reservation

	c.JSON(200, response.CommonResponse{
		Success: true,
	}.AddInterfaces(map[string]interface{}{
		"count":        len(reservations),
		"reservations": reservations,
	}))
}

func GetReservationHandler(c *gin.Context) {
	reservationId := c.Param("reservationId")
	reservation := model.Reservation{}

	if err := reservation.GetOne(map[string]interface{}{"id": reservationId}); err != nil {
		response.NotFound("Reservation not found").AbortWithError(c)
		return
	}

	userId := c.GetHeader("user_id")
	if userId != "" {
		// check if user is in participants or pending participants
		isParticipant := false
		for _, participant := range reservation.Participants {
			if participant == service.ParseToInt64(userId) {
				isParticipant = true
				break
			}
		}

		if !isParticipant {
			for _, pendingParticipant := range reservation.PendingParticipants {
				if pendingParticipant == service.ParseToInt64(userId) {
					isParticipant = true
					break
				}
			}
		}

		if !isParticipant {
			response.Forbidden("You are not a participant of this reservation").AbortWithError(c)
			return
		}
	}

	// TODO: call grpc to get room info

	c.JSON(200, response.CommonResponse{
		Success: true,
	}.AddInterfaces(map[string]interface{}{
		"reservation": reservation,
	}))
}

func CancelReservationHandler(c *gin.Context) {
	reservationId := c.Param("reservationId")
	reservation := model.Reservation{}

	if err := reservation.GetOne(map[string]interface{}{"id": reservationId}); err != nil {
		response.NotFound("Reservation not found").AbortWithError(c)
		return
	}

	userType := c.GetHeader("type")
	
	if(userType == "user") {
		userId := c.GetHeader("user_id")
		if userId != "" {
			// check if user is the first participant in the reservation
			if len(reservation.Participants) > 0 && reservation.Participants[0] != service.ParseToInt64(userId) {
				response.Forbidden("You are not allowed to cancel this reservation").AbortWithError(c)
				return
			}

			// current time must be at least 15 minutes before the reservation start time
			if !service.IsCurrentTimeBefore(reservation.StartDateTime, 15) {
				response.BadRequest("You can only cancel reservation 15 minutes before the start time").AbortWithError(c)
				return
			}
		}
	}

	reservation.Status = "canceled"

	if err := reservation.Update(); err != nil {
		response.InternalServerError("Failed to cancel reservation").AbortWithError(c)
		return
	}

	c.JSON(200, response.CommonResponse{
		Success: true,
	})
}

func ConfirmReservationHandler(c *gin.Context) {
	rcr := model.ReservationConfirmationRequest{}
	if err := c.ShouldBindJSON(&rcr); err != nil {
		response.BadRequest("Invalid request").AbortWithError(c)
		return
	}

	reservation := model.Reservation{}
	err := reservation.GetOne(map[string]interface{}{"id": rcr.ReservationId})
	if err != nil {
		response.NotFound("Reservation not found").AbortWithError(c)
		return
	}

	// check if user is in pending participants
	isPendingParticipant := false
	for i, pendingParticipant := range reservation.PendingParticipants {
		if pendingParticipant == int64(rcr.UserId) {
			isPendingParticipant = true
			reservation.PendingParticipants = append(reservation.PendingParticipants[:i], reservation.PendingParticipants[i+1:]...)
			break
		}
	}

	if !isPendingParticipant {
		response.Forbidden("You are already a participant of this reservation").AbortWithError(c)
	}

	// add user to participants
	reservation.Participants = append(reservation.Participants, int64(rcr.UserId))

	if len(reservation.PendingParticipants) == 0 {
		reservation.Status = "confirmed"
	}

	if err := reservation.Update(); err != nil {
		response.InternalServerError("Failed to confirm reservation").AbortWithError(c)
		return
	}

	c.JSON(200, response.CommonResponse{
		Success: true,
	})
}

func ApproveReservationHandler(c *gin.Context) {
	reservationId := c.Param("reservationId")
	reservation := model.Reservation{}

	if err := reservation.GetOne(map[string]interface{}{"id": reservationId}); err != nil {
		response.NotFound("Reservation not found").AbortWithError(c)
		return
	}

	approver := c.GetHeader("id")
	
	if(reservation.Status != "pending") {
		response.Forbidden("Reservation not confirmed").AbortWithError(c)
		return
	}
	if(approver == "") {
		response.Forbidden("You are not allowed to approve this reservation").AbortWithError(c)
		return
	}
	
	reservation.Approver = service.ParseToUint(approver)
	reservation.Status = "completed"
	if err := reservation.Update(); err != nil {
		response.InternalServerError("Failed to approve reservation").AbortWithError(c)
		return
	}

	c.JSON(200, response.CommonResponse{
		Success: true,
	})
	
}


