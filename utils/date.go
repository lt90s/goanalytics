package utils

import (
	"time"
)

var today time.Time

func init() {
	var err error

	if err != nil {
		panic(err)
	}
	now := time.Now()
	today = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	go updateToday()
}

func updateToday() {
	for {
		tomorrow := time.Date(today.Year(), today.Month(), today.Day()+1, 0, 0, 0, 0, today.Location())
		sleepSeconds := tomorrow.Unix() - time.Now().Unix()
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
		today = tomorrow
	}
}

func Now() time.Time {
	return time.Now()
}

func NowTimestamp() int64 {
	return time.Now().Unix()
}

func TimeToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func TimestampToDate(timestamp int64) time.Time {
	return TimeToDate(time.Unix(timestamp, 0))
}

func Today() time.Time {
	return today
}

func TodayTimestamp() int64 {
	return Today().Unix()
}

func TodayDiff(diff int) time.Time {
	return time.Date(today.Year(), today.Month(), today.Day()-diff, 0, 0, 0, 0, today.Location())
}
