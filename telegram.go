package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func telegramBot() {
	bot, err := tgbotapi.NewBotAPI(conf.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	sf := NewSFClient()

	getImg := func(query []string) (fileName string, err error) {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		card, err := findOneCard(query...)
		if err != nil {
			return "", err
		}

		fileName, err = sf.GetImage(ctx, card)
		if err != nil {
			return "", err
		}
		return fileName, nil
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "h":
			fallthrough
		case "help":
			msg.Text = "type /c[ard] <cardname> or /f[ind] <keywords>."
		case "card":
			fallthrough
		case "c":
			cmd := update.Message.CommandArguments()
			query := strings.Split(cmd, " - ")
			fmt.Printf("%v\n", query)
			fileName, err := getImg(query)
			if err != nil {
				log.Println(err)
				if err == context.DeadlineExceeded {
					msg.Text = "Timeout. Try again later."
				} else {
					msg.Text = "Not found."
				}
				bot.Send(msg)
				continue
			}
			cardImg := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, fileName)
			bot.Send(cardImg)
		case "find":
			fallthrough
		case "f":
			searchTerm := update.Message.CommandArguments()
			cards := searchCards(searchTerm)
			msg.Text = "Found:\n" + strings.Join(cards, "\n")
		default:
			msg.Text = "I don't know that command. Type /help."
		}

		if len(msg.Text) > 0 {
			bot.Send(msg)
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	}
}
