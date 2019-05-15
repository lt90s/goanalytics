package user

import (
	"errors"
	"github.com/lt90s/goanalytics/event/pubsub"
	"github.com/sirupsen/logrus"
)

type DailyScheduleEventData struct {
	Timestamp int64  `json:"timestamp"`
	AppId     string `json:"appIds"`
}

func dailyScheduleEventHandler(store Store) pubsub.EventHandler {
	return pubsub.EventHandlerFunc(func(data interface{}) error {
		eventData, ok := data.(*DailyScheduleEventData)
		if !ok {
			return errors.New("UserDailyScheduleEventHandler: data is not of type *UserDailyScheduleEventData")
		}
		calcDailyActiveNewUserPercent(eventData, store)
		calcDailyActiveUserAffinity(eventData, store)
		calcOpenAppCountDistribution(eventData, store)
		return nil
	})
}

func calculatePercent(a, b float64) float64 {
	if b == 0 {
		return 0
	}

	return a / b
}

func calcDailyActiveNewUserPercent(data *DailyScheduleEventData, store Store) {
	appId := data.AppId
	entry := logrus.WithFields(logrus.Fields{"timestamp": data.Timestamp, "appId": appId})
	dailyActiveCount, err := store.GetSimpleCPVSumTotal(appId, DailyActiveCPVCounter, data.Timestamp, data.Timestamp)
	if err != nil {
		entry.WithFields(logrus.Fields{"error": err.Error()}).Warn("[calcDailyActiveNewUserPercent] error")
		return
	}

	newUserCount, err := store.GetSimpleCPVSumTotal(appId, NewUserCPVCounter, data.Timestamp, data.Timestamp)
	if err != nil {
		entry.WithFields(logrus.Fields{"error": err.Error()}).Warn("[calcDailyActiveNewUserPercent] error")
		return
	}
	percent := calculatePercent(newUserCount, dailyActiveCount)
	store.AddSimpleCounter(appId, DailyActiveNewUserPercentSimpleCounter, data.Timestamp, percent)
}

func calcDailyActiveUserAffinity(data *DailyScheduleEventData, store Store) {
	timestamp := data.Timestamp

	saus := store.getUniqueActiveUserCount(data.AppId, timestamp-7*24*3600, timestamp)
	faus := store.getUniqueActiveUserCount(data.AppId, timestamp-15*24*3600, timestamp)
	taus := store.getUniqueActiveUserCount(data.AppId, timestamp-30*24*3600, timestamp)
	if saus == 0 || faus == 0 || taus == 0 {
		return
	}
	dailyActiveCount, _ := store.GetSimpleCPVSumTotal(data.AppId, DailyActiveCPVCounter, timestamp, timestamp)
	p1 := calculatePercent(dailyActiveCount, float64(saus))
	p2 := calculatePercent(dailyActiveCount, float64(faus))
	p3 := calculatePercent(dailyActiveCount, float64(taus))

	store.AddSlotCounter(data.AppId, DailyActiveUserAffinitySlotCounter, "7", timestamp, p1)
	store.AddSlotCounter(data.AppId, DailyActiveUserAffinitySlotCounter, "15", timestamp, p2)
	store.AddSlotCounter(data.AppId, DailyActiveUserAffinitySlotCounter, "30", timestamp, p3)
}

func calcOpenAppCountDistribution(data *DailyScheduleEventData, store Store) {
	store.calcOpenAppCountDistribution(data.AppId, data.Timestamp)
}
