package service

import (
	"github.com/FakJeongTeeNhoi/reservation-management/model"
	"github.com/go-co-op/gocron"
	"time"
)

var cron *gocron.Scheduler

func InitCron() {
	cron = gocron.NewScheduler(time.UTC)

	SetCronJob(deleteOldPendingReservations, 5*time.Minute)
}

func SetCronJob(job func(), interval time.Duration) {
	_, err := cron.Every(interval).Do(job)
	if err != nil {
		return
	}

	cron.StartAsync()
}

func deleteOldPendingReservations() {
	reservations := model.Reservations{}
	_ = reservations.GetOldPendingReservations(15 * time.Minute)

	// Delete all old pending reservations
	for _, reservation := range reservations {
		_ = reservation.Delete()
	}
}
