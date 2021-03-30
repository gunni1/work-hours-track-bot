package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/tucnak/telebot.v2"
	tbApi "gopkg.in/tucnak/telebot.v2"
	"strconv"
	"strings"
	"time"
)

const (
	DB_NAME     = "WORK_DB"
	DB_COL_NAME = "WORK_LOG"
)

type BotContext struct {
	DbSession *mgo.Session
	Bot       *telebot.Bot
}

func (ctx BotContext) RegisterCommands() {
	bot := ctx.Bot
	bot.Handle("/start", func(m *tbApi.Message) {
		bot.Send(m.Sender, "WorkHoursTrackBot is a bot about tracking your work time account. "+
			"Type /help to see usage examples.")
	})
	bot.Handle("/help", func(m *tbApi.Message) {
		bot.Send(m.Sender, "/initovertime 6,25 \nInit overtime account with 6 hours, 15 minutes")
		bot.Send(m.Sender, "/today 7 \nSet 7 hours as today's work time")
	})
	bot.Handle("/initovertime", func(m *tbApi.Message) {
		//TODO: nur den ersten token aus m.Payload übergeben
		workload, parseErr := ParseWorkTimeFloat(m.Payload)
		if parseErr != nil {
			bot.Send(m.Sender, fmt.Sprintf("%s is not a valid floatig point number.", m.Payload))
			return
		}
		//TODO: Wenn bereits existiert, alle vorherigen löschen
		worklog := WorkLog{
			UserId:    m.Sender.ID,
			WorkLoad:  workload,
			TimeStamp: time.Now(),
			IsInitial: true,
		}
		dbSession := ctx.DbSession.Clone()
		defer dbSession.Close()
		workLogs := dbSession.DB(DB_NAME).C(DB_COL_NAME)
		insErr := workLogs.Insert(worklog)
		if insErr != nil {
			bot.Send(m.Sender, "Initialization failed due internal error. Sry :(")
		} else {
			bot.Send(m.Sender, fmt.Sprintf("Initialized with %.2f hours.", workload))
		}
	})
}

func ParseWorkTimeFloat(input string) (float64, error) {
	unified := strings.ReplaceAll(input, ",", ".")
	result, err := strconv.ParseFloat(unified, 64)
	return result, err
}
