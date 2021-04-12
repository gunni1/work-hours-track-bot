package main

import (
	"gopkg.in/mgo.v2"
	tbApi "gopkg.in/tucnak/telebot.v2"
	"log"
	"os"
	"time"
)

func main() {
	token := parseEnvMandatory("BOT_TOKEN")
	dbURL := parseEnvMandatory("DB_URL")

	bot, err := tbApi.NewBot(tbApi.Settings{
		Token:  token,
		Poller: &tbApi.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	session, err := mgo.Dial(dbURL)
	if err != nil {
		log.Fatalf("Unable to establish DB connection to %s: %s", dbURL, err)
	}
	defer session.Close()

	context := BotContext{
		DbSession: session,
		Bot:       bot,
	}
	context.RegisterCommands()
	log.Println("Bot Ready.")
	bot.Start()
}

func parseEnvMandatory(variableKey string) string {
	variableValue := os.Getenv(variableKey)
	if variableValue == "" {
		log.Fatalln("Environment variable: " + variableKey + " is empty")
	}
	return variableValue
}
