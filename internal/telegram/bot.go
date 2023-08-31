package telegram

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"

	"github.com/parsaaes/rating-telegram-bot/internal/model"

	"github.com/forPelevin/gomoji"
)

type Bot struct {
	bot *tgbotapi.BotAPI
	categoryRepo model.CategoryRepo
	itemRepo model.ItemRepo
	rateRepo model.RateRepo
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
		bot: bot,
		categoryRepo: categoryRepo,
		itemRepo: itemRepo,
		rateRepo: rateRepo,
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
					_ = b.replay("invalid format. the correct format is " + create, update)

					continue
				}

				icon := ""

				if len(argsByDash) > 1  {
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
					_ = b.replay("invalid format. the correct format is " + add, update)

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
			case "rate":
				if len(argsByDash) < 2 {
					_ = b.replay("invalid format. the correct format is " + rate, update)

					continue
				}

				itemName := argsByDash[0]
				scoreStr := argsByDash[1]

				score, err := strconv.ParseFloat(scoreStr, 64)
				if err != nil {
					logrus.Errorf("error converting the rating string: %s", err.Error())

					continue
				}

				if !((0 <= score && score <= 5) && math.Mod(score, 0.25) == 0) {
					_ = b.replay("Your score should be a multiple of 0.25 between 0 and 5.", update)

					continue
				}

				itemID, err := b.itemRepo.FindIDByTitleAndGroupID(itemName, groupID)
				if err != nil {
					logrus.Errorf("error getting the item id: %s", err.Error())

					continue
				}

				if err := b.rateRepo.Save(&model.Rate{
					Rater:  update.Message.From.UserName,
					ItemID: itemID,
					Rate:   score,
				}); err != nil {
					logrus.Errorf("error saving the rate: %s", err.Error())

					continue
				}

				_ = b.replay(fmt.Sprintf("✅ You rated %s:\n%s from 5.",
					itemName, strconv.FormatFloat(score, 'f', 2, 64)), update)
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
				msg := fmt.Sprintf("👋 *here's the list of commands*:\n\n%s - *create a new category*\n\n%s - *add a new title to a category*\n\n%s - *rate a title*\n\n%s - *list all items*\n",
					create, add, rate, list)

				_ = b.replay(msg, update)
			}
		}
	}
}

func (b *Bot) replay (text string, update tgbotapi.Update) error {
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