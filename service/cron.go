package service

import (
	"github.com/FakJeongTeeNhoi/reservation-management/model"
	"github.com/go-co-op/gocron"
	"time"
)

var cron *gocron.Scheduler

func InitCron() {
	cron = gocron.NewScheduler(time.UTC)

	SetCronJob(cancelOldPendingReservations, 5*time.Minute)
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
	_ = reservations.GetOldPendingReservations(15 * time.Minute)

	// Delete all old pending reservations
	for _, reservation := range reservations {
		reservation.Status = "cancelled"
		_ = reservation.Update()
	}
}
