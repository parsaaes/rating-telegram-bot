package main

import (
	"log"

	"github.com/parsaaes/rating-telegram-bot/internal/config"
	"github.com/parsaaes/rating-telegram-bot/internal/database"
	"github.com/parsaaes/rating-telegram-bot/internal/model"
	"github.com/parsaaes/rating-telegram-bot/internal/telegram"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		logrus.Fatal("error reading configs: %s", err.Error())

	}

	db, err := database.New("rating.db")
	if err != nil {
		logrus.Fatal("error connecting to db: %s", err.Error())
	}

	if cfg.Debug {
		db = db.Debug()

		logrus.SetLevel(logrus.DebugLevel)
	}

	if err := db.AutoMigrate(model.Category{}, model.Item{}, model.Rate{}); err != nil {
		logrus.Fatal("error running migrations: %s", err.Error())
	}

	categoryRepo := &model.SQLCategoryRepo{DB: db}
	itemRepo := &model.SQLItemRepo{DB: db}
	rateRepo := &model.SQLRateRepo{DB: db}

	bot, err := telegram.New(cfg.Token, categoryRepo, itemRepo, rateRepo)
	if err != nil {
		log.Panic(err)
	}

	bot.Run()
}
