package main

import "time"

type WorkLog struct {
	UserId    int       `bson:"userId"`
	WorkLoad  float64   `bson:"workLoad"`
	TimeStamp time.Time `bson:"timeStamp"`
}

type Account struct {
	UserId                 int     `bson:"userId"`
	WeekWorkHours          float64 `bson:"weekWorkHours"`
	InitialWorkHourBalance float64 `bson:"initialWorkHourBalance"`
}
