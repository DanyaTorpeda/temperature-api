package main

import (
	"os"
	weatherapi "third-party-api"
	"third-party-api/pkg/handler"
	"third-party-api/pkg/service"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logrus.Errorf("error occured loading env variables: %s", err.Error())
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: os.Getenv("PASSWORD"),
		DB:       0,
		Protocol: 2,
	})

	services := service.NewService(redisClient)
	handlers := handler.NewHandler(services)
	srv := new(weatherapi.Server)
	if err := srv.Run("8080", handlers.InitRoutes()); err != nil {
		logrus.Errorf("error occured running server: %s", err.Error())
	}
}
