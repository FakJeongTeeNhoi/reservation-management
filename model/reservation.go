package model

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
	"time"
)

type Reservation struct {
	gorm.Model
	Participants        pq.Int64Array `json:"participants" gorm:"type:BIGINT[]"`
	PendingParticipants pq.Int64Array `json:"pending_participants" gorm:"type:BIGINT[]"`
	Status              string        `json:"status"`
	Approver            uint          `json:"approver"`
	StartDateTime       string        `json:"start_date_time"`
	EndDateTime         string        `json:"end_date_time"`
	RoomId              uint          `json:"room_id"`
}

type Reservations []Reservation

type ReservationCreateRequest struct {
	PendingParticipants pq.Int64Array `json:"pending_participants"`
	StartDateTime       string        `json:"start_date_time"`
	EndDateTime         string        `json:"end_date_time"`
	RoomId              uint          `json:"room_id"`
}

type ReservationConfirmationRequest struct {
	UserId        uint `json:"user_id"`
	ReservationId uint `json:"reservation_id"`
}

func (r *ReservationCreateRequest) ToReservation() Reservation {
	return Reservation{
		PendingParticipants: r.PendingParticipants,
		StartDateTime:       r.StartDateTime,
		EndDateTime:         r.EndDateTime,
		RoomId:              r.RoomId,
	}
}

func (r *Reservation) Create() error {
	result := MainDB.Model(&Reservation{}).Create(r)
	return result.Error
}

func (r *Reservation) GetOne(filter interface{}) error {
	result := MainDB.Model(&Reservation{}).Where(filter).First(r)
	return result.Error
}

func (r *Reservations) GetAll(filter interface{}) error {
	result := MainDB.Model(&Reservation{}).Where(filter).Find(r)
	return result.Error
}

func (r *Reservations) GetByRoomIdAndTimePeriod(roomId uint, start string, end string) error {
	result := MainDB.Model(&Reservation{}).Where("room_id = ?", roomId).
		Where("start_date_time <= ? AND end_date_time >= ?", start, start).
		Or("start_date_time <= ? AND end_date_time >= ?", end, end).
		Find(r)
	return result.Error
}

func (r *Reservation) Update() error {
	result := MainDB.Model(&Reservation{}).Where("id = ?", r.ID).Updates(r)
	return result.Error
}

func (r *Reservation) Delete() error {
	result := MainDB.Model(&Reservation{}).Where("id = ?", r.ID).Delete(
		map[string]interface{}{
			"id": r.ID,
		})
	return result.Error
}

func (r *Reservations) GetOldPendingReservations(period time.Duration) error {
	result := MainDB.Model(&Reservation{}).
		Where("status = ?", "pending").
		Where("created_at <= ?", time.Now().Add(-period)).
		Find(r)
	return result.Error
}
