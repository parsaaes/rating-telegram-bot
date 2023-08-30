package main

import (
	"github.com/parsaaes/rating-telegram-bot/internal/database"
	"github.com/parsaaes/rating-telegram-bot/internal/model"

	"github.com/sirupsen/logrus"
)

func main() {
	db, err := database.New("rating.db")

	if err != nil {
		logrus.Fatal("error connecting to db: %s", err.Error())
	}

	if err := db.AutoMigrate(model.Category{}); err != nil {
		logrus.Fatal("error running migrations: %s", err.Error())
	}
}
