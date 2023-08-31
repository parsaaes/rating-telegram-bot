package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/forPelevin/gomoji"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"

	"github.com/parsaaes/rating-telegram-bot/internal/model"
)

type Bot struct {
	bot          *tgbotapi.BotAPI
	categoryRepo model.CategoryRepo
	itemRepo     model.ItemRepo
	rateRepo     model.RateRepo
}

func New(
	token string,
	categoryRepo model.CategoryRepo,
	itemRepo model.ItemRepo,
	rateRepo model.RateRepo,
) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		bot:          bot,
		categoryRepo: categoryRepo,
		itemRepo:     itemRepo,
		rateRepo:     rateRepo,
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

			groupID := strconv.FormatInt(update.Message.Chat.ID, 10)

			rawArgs := update.Message.CommandArguments()

			rawArgsByDash := strings.Split(rawArgs, "-")
			argsByDash := []string{}

			for i := range rawArgsByDash {
				if rawArgsByDash[i] != "" {
					argsByDash = append(argsByDash, strings.TrimSpace(rawArgsByDash[i]))
				}
			}

			switch update.Message.Command() {
			case "create":
				if len(argsByDash) < 1 {
					_ = b.replay("invalid format. the correct format is "+create, update)

					continue
				}

				icon := ""

				if len(argsByDash) > 1 {
					emojies := gomoji.FindAll(strings.TrimSpace(argsByDash[1]))
					if len(emojies) > 0 {
						icon = emojies[0].Character
					}
				}

				if err := b.categoryRepo.Create(&model.Category{
					GroupID: groupID,
					Name:    argsByDash[0],
					Icon:    icon,
				}); err != nil {
					logrus.Errorf("error creating category: %s", err.Error())

					continue
				}
			case "add":
				if len(argsByDash) < 2 {
					_ = b.replay("invalid format. the correct format is "+add, update)

					continue
				}

				categoryName := argsByDash[0]
				title := argsByDash[1]

				title = strings.TrimSpace(title)

				category, err := b.categoryRepo.FindByName(categoryName, groupID)
				if err != nil {
					logrus.Errorf("error getting category: %s", err.Error())

					continue
				}

				if err := b.itemRepo.Create(&model.Item{
					Title:      title,
					CategoryID: category.ID,
				}); err != nil {
					logrus.Errorf("error saving item: %s", err.Error())

					continue
				}

				_ = b.replay(fmt.Sprintf("✅\n %s added to %s.", title, categoryName), update)
			case "list":
				list, err := b.rateRepo.List(groupID)
				if err != nil {
					logrus.Errorf("error getting items: %s", err.Error())

					continue
				}

				msg := ""

				for category, items := range list {
					msg += category.Icon + " <b>" + category.Name + "</b>" + ":\n"

					for item, rates := range items {
						msg += item.Title + ":\n"

						for _, rate := range rates {
							msg += "@" + rate.Rater + " => " +
								strconv.FormatFloat(rate.Rate, 'f', 2, 64) + " out of 5\n"
						}
					}

					msg += "\n"
				}

				msg = strings.TrimSpace(msg)

				var msgToSend tgbotapi.MessageConfig

				if msg != "" {
					msgToSend = tgbotapi.NewMessage(update.Message.Chat.ID, msg)
				} else {
					msgToSend = tgbotapi.NewMessage(update.Message.Chat.ID, "no categories found. create one first.")
				}

				msgToSend.ParseMode = "html"

				if _, err := b.bot.Send(msgToSend); err != nil {
					logrus.Errorf("error sending list message: %s", err.Error())

					continue
				}

			case "help":
				msg := fmt.Sprintf("👋 *here's the list of commands*:\n\n%s - *create a new category*\n\n%s - *add a new title to a category*\n\n%s - *list titles to score*\n\n%s - *list rates*\n",
					create, add, titles, list)

				_ = b.replay(msg, update)

			case "titles":
				list, err := b.rateRepo.List(groupID)
				if err != nil {
					logrus.Errorf("error getting items: %s", err.Error())

					continue
				}

				var rows [][]tgbotapi.InlineKeyboardButton

				for category, items := range list {
					for item := range items {
						titleForButton := category.Name+" "+item.Title

						if len(titleForButton) > 25 {
							titleForButton = string([]rune(titleForButton)[:20]) + "..."
						}

						rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(titleForButton, requestRateKeyboardCallback+strconv.FormatUint(uint64(item.ID), 10)+":"+titleForButton)))
					}
				}

				if len(rows) == 0 {
					_ = b.replay("no title found.", update)
				}

				var keyboard = tgbotapi.NewInlineKeyboardMarkup(
					rows...,
				)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "list of titles:")

				msg.ReplyMarkup = keyboard

				if _, err := b.bot.Send(msg); err != nil {
					logrus.Errorf("error sending titles keyboard: %s", err.Error())
				}
			}
		} else if update.CallbackQuery != nil {
			if strings.HasPrefix(update.CallbackQuery.Data, requestRateKeyboardCallback) {
				callBackDataArgs := strings.Split(strings.TrimPrefix(update.CallbackQuery.Data, requestRateKeyboardCallback), ":")

				if len(callBackDataArgs) != 2 {
					logrus.Errorf("callback data args are not correct: %s", update.CallbackQuery.Data)

					_ = b.buttonFailed(update)

					continue
				}

				itemIDStr := callBackDataArgs[0]
				name := callBackDataArgs[1]

				itemID, err := strconv.Atoi(itemIDStr)
				if err != nil {
					logrus.Errorf("error getting item id from callback: %s", err.Error())

					_ = b.buttonFailed(update)

					continue
				}

				var rows [][]tgbotapi.InlineKeyboardButton

				for i := 0; i < 5; i++ {
					var row []tgbotapi.InlineKeyboardButton

					for _, p := range []string{"", ".25", ".5", ".75"} {
						score := strconv.Itoa(i) + p

						row = append(row, tgbotapi.NewInlineKeyboardButtonData(score, rateCallbackFmt(score, itemID)))
					}

					rows = append(rows, row)
				}

				rows = append(rows, []tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData("5", rateCallbackFmt("5", itemID)),
				})

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "choose your score for "+name+":")

				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "choose your score")
				if _, err := b.bot.Request(callback); err != nil {
					logrus.Errorf("error responding to callback: %s", err.Error())

					continue
				}

				if _, err := b.bot.Send(msg); err != nil {
					logrus.Errorf("error sending score keyboard: %s", err.Error())

					continue
				}
			} else if strings.HasPrefix(update.CallbackQuery.Data, rateCallback) {
				callbackDataArgs := strings.Split(strings.TrimPrefix(update.CallbackQuery.Data, rateCallback), ":")

				if len(callbackDataArgs) != 2 {
					logrus.Errorf("error in callback data args: %s", update.CallbackQuery.Data)

					_ = b.buttonFailed(update)

					continue
				}

				scoreStr := callbackDataArgs[0]
				itemIDStr := callbackDataArgs[1]

				score, err := strconv.ParseFloat(scoreStr, 64)
				if err != nil {
					logrus.Errorf("error converting score from callback: %s", err.Error())

					_ = b.buttonFailed(update)

					continue
				}

				itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
				if err != nil {
					logrus.Errorf("error converting id from callback: %s", err.Error())

					_ = b.buttonFailed(update)

					continue
				}

				if err := b.rateRepo.Save(&model.Rate{
					Rater:  update.CallbackQuery.From.UserName,
					ItemID: uint(itemID),
					Rate:   score,
				}); err != nil {
					logrus.Errorf("error saving rate: %s", err.Error())

					_ = b.buttonFailed(update)

					continue
				}

				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "rate saved.")
				if _, err := b.bot.Request(callback); err != nil {
					logrus.Errorf("error responding to callback: %s", err.Error())

					continue
				}
			}
		}
	}
}

func (b *Bot) buttonFailed(update tgbotapi.Update) error {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "failed")
	if _, err := b.bot.Request(callback); err != nil {
		logrus.Errorf("error responding to callback: %s", err.Error())

		return err
	}

	return nil
}

func (b *Bot) replay(text string, update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf(text))

	msg.ReplyToMessageID = update.Message.MessageID

	msg.ParseMode = "markdown"

	if _, err := b.bot.Send(msg); err != nil {
		logrus.Errorf("error replaying: %s", err.Error())

		return err
	}

	return nil
}
