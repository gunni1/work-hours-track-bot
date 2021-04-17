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

var answerState = make(map[int]string)
var accWiz = make(map[int]Account)

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
		bot.Send(m.Sender, "/balance \nSee your total overtime balance")

	})
	bot.Handle("/today", func(m *tbApi.Message) {
		workload, parseErr := ParseFloat(m.Payload)
		if parseErr != nil {
			bot.Send(m.Sender, parseErr.Error())
			return
		}
		workLog := WorkLog{
			UserId:    m.Sender.ID,
			WorkLoad:  workload,
			TimeStamp: time.Now(),
		}
		dbSession := ctx.DbSession.Clone()
		defer dbSession.Close()
		workLogs := dbSession.DB(DB_NAME).C(DB_COL_WORK)

		existCount, err := workLogs.Find(bson.M{"userId": m.Sender.ID, "timeStamp": bson.M{
			"$gt": TodayBegin(),
			"$lt": TodayEnd(),
		}}).Count()
		if err == nil && existCount > 0 {
			removeErr := workLogs.Remove(bson.M{"userId": m.Sender.ID, "timeStamp": bson.M{
				"$gt": TodayBegin(),
				"$lt": TodayEnd(),
			}})
			if removeErr != nil {
				log.Println(removeErr.Error())
			}
			log.Printf("Removed todays worklog for user: %d", m.Sender.ID)
		}

		insErr := workLogs.Insert(workLog)
		if insErr != nil {
			bot.Send(m.Sender, "Save Work failed due internal error. Sry :(")
		} else {
			bot.Send(m.Sender, fmt.Sprintf("Worked %.2f hours.", workload))
		}
	})
	bot.Handle("/balance", func(m *tbApi.Message) {
		dbSession := ctx.DbSession.Clone()
		defer dbSession.Close()
		workLogs := dbSession.DB(DB_NAME).C(DB_COL_WORK)
		accounts := dbSession.DB(DB_NAME).C(DB_COL_ACC)
		var usersAcc Account
		accFindErr := accounts.Find(bson.M{"userId": m.Sender.ID}).One(&usersAcc)
		if accFindErr != nil {
			bot.Send(m.Sender, "You have no work log account yet. Please start with /init")
			return
		}
		var userWorkLogs []WorkLog
		workLogs.Find(bson.M{"userId": m.Sender.ID}).All(&userWorkLogs)
		workLogBalance := CalculateWorkDayBalance(userWorkLogs, usersAcc.WeekWorkHours)
		totalBalance := workLogBalance + usersAcc.InitialWorkHourBalance
		bot.Send(m.Sender, fmt.Sprintf("Your Current Work Balance is: %.2f hours", totalBalance))
	})
	//Interactive Initialization
	bot.Handle("/init2", func(m *tbApi.Message) {
		bot.Send(m.Sender, "Please give me your current overtime balance.")
		answerState[m.Sender.ID] = "ASKED_OT_BALANCE"
	})
	bot.Handle(tbApi.OnText, func(m *tbApi.Message) {

		if answerState[m.Sender.ID] == "ASKED_OT_BALANCE" {
			overtimeBalance, otpErr := ParseFloat(m.Text)
			if otpErr != nil {
				bot.Send(m.Sender, "Please type your overtime balance in hours as a number. For example '4,75' for 4 hours and 45 minutes")
				return
			}
			newAccount := Account{
				UserId:                 m.Sender.ID,
				InitialWorkHourBalance: overtimeBalance,
			}
			accWiz[m.Sender.ID] = newAccount
			bot.Send(m.Sender, fmt.Sprintf("Ok, you start with %.2f hours overtime.", overtimeBalance))
			bot.Send(m.Sender, "How many hours per week do you work?")
			answerState[m.Sender.ID] = "ASKED_WEEK_HOURS"
		} else if answerState[m.Sender.ID] == "ASKED_WEEK_HOURS" {
			weekWorkHours, wpErr := ParseFloat(m.Text)
			if wpErr != nil {
				bot.Send(m.Sender, "Please type your weekly work hours as a number. For example '40'")
				return
			}
			dbSession := ctx.DbSession.Clone()
			defer dbSession.Close()
			accounts := dbSession.DB(DB_NAME).C(DB_COL_ACC)
			account := accWiz[m.Sender.ID]
			account.WeekWorkHours = weekWorkHours
			//Remove if already exists
			findErr := accounts.Find(bson.M{"userId": m.Sender.ID})
			if findErr == nil {
				accounts.Remove(bson.M{"userId": m.Sender.ID})
				log.Printf("Removed account balance for user: %d", m.Sender.ID)
			}
			accounts.Insert(account)
			delete(accWiz, m.Sender.ID)
			delete(answerState, m.Sender.ID)
			bot.Send(m.Sender, fmt.Sprintf("Account initialized with %.2f hours overtime and %.2f weekly work time.",
				account.InitialWorkHourBalance, account.WeekWorkHours))
			bot.Send(m.Sender, "Use eg. '/today 8' to log your daily work time")
		}
	})
}

func CalculateWorkDayBalance(workLogs []WorkLog, weekWorkHours float64) float64 {
	actualSum := 0.0
	targetHours := (weekWorkHours / 5) * float64(len(workLogs))
	for _, workLog := range workLogs {
		actualSum += workLog.WorkLoad
	}
	return actualSum - targetHours
}

func TodayBegin() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func TodayEnd() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
}

func ParseFloat(input string) (float64, error) {
	unified := strings.ReplaceAll(input, ",", ".")
	result, parseErr := strconv.ParseFloat(unified, 64)
	if parseErr != nil {
		return result, errors.New(fmt.Sprintf("%s is not a valid floatig point number.", input))
	}
	return result, nil
}
