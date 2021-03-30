package main

import "time"

type WorkLog struct {
	UserId    int       `bson:"userId"`
	WorkLoad  float64   `bson:"workLoad"`
	TimeStamp time.Time `bson:"timeStamp"`
	IsInitial bool      `bson:"isInitial"`
}
