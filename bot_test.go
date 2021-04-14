package main

import (
	"testing"
	"time"
)

func TestParseWorkTimeFloat(t *testing.T) {
	testsets := []struct {
		Input    string
		ExpFloat float64
	}{
		{"0", 0},
		{"1,5", 1.5},
		{"1,7500000", 1.75},
		{"10.25", 10.25},
	}
	for _, testset := range testsets {
		result, err := ParseFloat(testset.Input)
		if err != nil {
			t.Error(err)
		}
		if testset.ExpFloat != result {
			t.Errorf("expected: %f to be equal to result: %f", testset.ExpFloat, result)
		}
	}
}

func TestCalculateWorkDayBalance(t *testing.T) {
	testSets := []struct {
		WorkLogs      []WorkLog
		WeekWorkHours float64
		ExpResult     float64
	}{
		{[]WorkLog{{UserId: 1, TimeStamp: time.Now(), WorkLoad: 7}}, 40, -1},
		{[]WorkLog{}, 40, 0},
		{[]WorkLog{
			{UserId: 1, TimeStamp: time.Now(), WorkLoad: 9},
			{UserId: 1, TimeStamp: time.Now(), WorkLoad: 9},
		}, 40, 2},
		{[]WorkLog{{UserId: 1, TimeStamp: time.Now(), WorkLoad: 7}}, 35, 0},
		{[]WorkLog{
			{UserId: 1, TimeStamp: time.Now(), WorkLoad: 8},
			{UserId: 1, TimeStamp: time.Now(), WorkLoad: 8},
			{UserId: 1, TimeStamp: time.Now(), WorkLoad: 8},
			{UserId: 1, TimeStamp: time.Now(), WorkLoad: 6.5},
			{UserId: 1, TimeStamp: time.Now(), WorkLoad: 7.25},
		}, 38, -0.25},
	}
	for i, testSet := range testSets {
		result := CalculateWorkDayBalance(testSet.WorkLogs, testSet.WeekWorkHours)
		if testSet.ExpResult != result {
			t.Errorf("%d - expected: %f to be equal to %f", i, testSet.ExpResult, result)
		}
	}
}

func TestCalculateWorkDayBalanceErrorInput(t *testing.T) {
	var workLogs []WorkLog
	result := CalculateWorkDayBalance(workLogs, 40)
	if result != 0 {
		t.Error("Empty work log list did not work")
	}
}
