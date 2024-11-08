package main

import (
	"fmt"
	"github.com/FakJeongTeeNhoi/reservation-management/model"
	"github.com/FakJeongTeeNhoi/reservation-management/router"
	"github.com/FakJeongTeeNhoi/reservation-management/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting server...")

	// TODO: Connect to database using gorm
	model.InitDB()

	service.ConnectMailer()

	service.InitCron()

	server := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}

	server.Use(cors.New(corsConfig))

	api := server.Group("/api")

	// TODO: Add routes here
	router.ReserveRouterGroup(api)

	err = server.Run(":" + os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}

	// TODO: Add graceful shutdown
}
