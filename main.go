package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"os"
	"time"
)

func main() {
	token := parseEnvMandatory("BOT_TOKEN")

	bot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	bot.Handle("/hello", func(m *tb.Message) {
		bot.Send(m.Sender, "Hello World!!!!")
	})

	bot.Start()
}

func parseEnvMandatory(variableKey string) string {
	variableValue := os.Getenv(variableKey)
	if variableValue == "" {
		log.Fatalln("Environment variable: " + variableKey + " is empty")
	}
	return variableValue
}
