package service

import (
	"github.com/FakJeongTeeNhoi/reservation-management/controller"
	"github.com/go-co-op/gocron"
	"time"
)

var cron *gocron.Scheduler

func InitCron() {
	cron = gocron.NewScheduler(time.UTC)

	SetCronJob(controller.DeleteOldPendingReservations, 5*time.Minute)
}

func SetCronJob(job func(), interval time.Duration) {
	_, err := cron.Every(interval).Do(job)
	if err != nil {
		return
	}

	cron.StartAsync()
}
