package service

import (
	"github.com/FakJeongTeeNhoi/reservation-management/model"
	"github.com/go-co-op/gocron"
	"time"
)

var cron *gocron.Scheduler

func InitCron() {
	cron = gocron.NewScheduler(time.UTC)

	SetCronJob(cancelOldPendingReservations, 1*time.Minute)
	SetCronJob(cancelLateReservations, 1*time.Minute)
	SetCronJob(requestReservations, 1*time.Minute)
}

func SetCronJob(job func(), interval time.Duration) {
	_, err := cron.Every(interval).Do(job)
	if err != nil {
		return
	}

	cron.StartAsync()
}

func cancelOldPendingReservations() {
	reservations := model.Reservations{}
	_ = reservations.GetOldPendingReservations(30 * time.Minute) 

	// Delete all old pending reservations
	for _, reservation := range reservations {
		reservation.Status = "canceled"
		_ = reservation.Update()
	}
}

func cancelLateReservations() {
	reservations := model.Reservations{}
	_ = reservations.GetLateReservations(15 * time.Minute) 

	for _, reservation := range reservations {
		reservation.Status = "canceled"
		_ = reservation.Update()
	}
}

func requestReservations() {
	reservations := model.Reservations{}
	_ = reservations.GetConfirmedReservations(15 * time.Minute)

	for _, reservation := range reservations {
		reservation.Status = "pending"
		_ = reservation.Update()
	}
}

