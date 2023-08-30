package telegram

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"

	"github.com/parsaaes/rating-telegram-bot/internal/model"
)

type Bot struct {
	bot *tgbotapi.BotAPI
	categoryRepo model.CategoryRepo
}

func New(token string, categoryRepo model.CategoryRepo) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		bot: bot,
		categoryRepo: categoryRepo,
	}, nil
}

func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			logrus.Debugf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if !update.Message.Chat.IsGroup() {
				continue
			}

			if !update.Message.IsCommand() {
				continue
			}

			switch update.Message.Command() {
			case "create":
				rawArgs := update.Message.CommandArguments()

				if len(rawArgs) < 1 {
					continue
				}

				if err := b.categoryRepo.Create(&model.Category{
					GroupID: strconv.FormatInt(update.Message.Chat.ID, 10),
					Name:    rawArgs,
				}); err != nil {
					logrus.Errorf("error creating category: %s", err.Error())
				}
			}
		}
	}
}