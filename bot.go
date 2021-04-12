package main

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/tucnak/telebot.v2"
	tbApi "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	DB_NAME     = "WORK_DB"
	DB_COL_WORK = "WORK_LOG"
	DB_COL_ACC  = "WORK_ACC"
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
		bot.Send(m.Sender, "/init 6,25 40 \nInit overtime account with 6 hours, 15 minutes and 40 hours of work per day")
		bot.Send(m.Sender, "/today 7 \nSet 7 hours as today's work time")
	})
	bot.Handle("/init", func(m *tbApi.Message) {
		//TODO: Überarbeiten. Besser eigene collection mit initialwert und wochenarbeitszeit
		args := strings.Split(m.Payload, " ")
		initial, wlErr := ParseFloat(args[0])
		workHours, whErr := ParseFloat(args[1])
		if wlErr != nil || whErr != nil {
			bot.Send(m.Sender, fmt.Sprintf("%s contains a invalid floatig point number.", m.Payload))
			return
		}

		newAccount := Account{
			UserId:                 m.Sender.ID,
			InitialWorkHourBalance: initial,
			WeekWorkHours:          workHours,
		}

		//TODO: Wenn bereits existiert, alle vorherigen löschen
		dbSession := ctx.DbSession.Clone()
		defer dbSession.Close()
		accounts := dbSession.DB(DB_NAME).C(DB_COL_ACC)
		var sendersAcc []Account
		findErr := accounts.Find(bson.M{"userId": m.Sender.ID}).All(&sendersAcc)
		if findErr == nil {
			accounts.Remove(bson.M{"userId": m.Sender.ID})
			log.Printf("Removed account balance for user: %d", m.Sender.ID)
		}

		insErr := accounts.Insert(newAccount)
		if insErr != nil {
			bot.Send(m.Sender, "Initialization failed due internal error. Sry :(")
		} else {
			bot.Send(m.Sender, fmt.Sprintf("Initialized overtime balance with %.2f hours and "+
				"%.1f work hours per week", initial, workHours))
		}
	})
	bot.Handle("/today", func(m *tbApi.Message) {
		//TODO: nur den ersten token aus m.Payload übergeben
		workload, parseErr := ParseFloat(m.Payload)
		if parseErr != nil {
			bot.Send(m.Sender, parseErr.Error())
			return
		}
		worklog := WorkLog{
			UserId:    m.Sender.ID,
			WorkLoad:  workload,
			TimeStamp: time.Now(),
		}
		dbSession := ctx.DbSession.Clone()
		defer dbSession.Close()
		workLogs := dbSession.DB(DB_NAME).C(DB_COL_WORK)
		//TODO: Besteht für heute bereits ein Eeintrag? Dann diesen löschen
		insErr := workLogs.Insert(worklog)
		if insErr != nil {
			bot.Send(m.Sender, "Save Work failed due internal error. Sry :(")
		} else {
			bot.Send(m.Sender, fmt.Sprintf("Worked %.2f hours.", workload))
		}
	})
	bot.Handle("/overtime", func(m *tbApi.Message) {
		//TODO: Alle Worklogs laden, Wochenarbeitszeit laden, zu gesamtsado verrechnen
	})
}

func ParseFloat(input string) (float64, error) {
	unified := strings.ReplaceAll(input, ",", ".")
	result, parseErr := strconv.ParseFloat(unified, 64)
	if parseErr != nil {
		return result, errors.New(fmt.Sprintf("%s is not a valid floatig point number.", input))
	}
	return result, nil
}
