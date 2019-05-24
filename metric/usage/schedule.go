package usage

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
		entry := logrus.WithFields(logrus.Fields{"data": data})
		if !ok {
			return errors.New("UserDailyScheduleEventHandler: data is not of type *UserDailyScheduleEventData")
		}

		err := calculateEachUsageAverageTime(eventData, store)
		if err != nil {
			entry.Warn("calculateEachUsageAverageTime error: ", err.Error())
		}

		err = calculateDailyUsageAverageTime(eventData, store)
		if err != nil {
			entry.Warn("calculateDailyUsageAverageTime error: ", err.Error())
		}

		err = calculateDailyUsageTimeDistribution(eventData, store)
		if err != nil {
			entry.Warn("calculateDailyUsageTimeDistribution error: ", err.Error())
		}

		return err
	})
}

func calculateEachUsageAverageTime(data *DailyScheduleEventData, store Store) error {
	totalTime, err := store.GetSimpleCounterSum(data.AppId, UsageTimeTotalSimpleCounter, data.Timestamp, data.Timestamp)
	if err != nil {
		return err
	}
	totalCount, err := store.GetSimpleCounterSum(data.AppId, UsageSimpleCounter, data.Timestamp, data.Timestamp)
	if err != nil {
		return err
	}

	if totalCount == 0 {
		return nil
	}
	return store.SetSimpleCounter(data.AppId, EachUsageAverageTimeSimpleCounter, data.Timestamp, totalTime/totalCount)
}

func calculateDailyUsageAverageTime(data *DailyScheduleEventData, store Store) error {
	totalTime, err := store.getTotalUsageTime(data.AppId, data.Timestamp)
	if err != nil {
		return err
	}

	totalCount, err := store.getDeviceCount(data.AppId, data.Timestamp)
	if err != nil {
		return err
	}

	if totalCount == 0 {
		return nil
	}

	return store.SetSimpleCounter(data.AppId, DailyUsageAverageTimeSimpleCounter, data.Timestamp, totalTime/float64(totalCount))
}

func calculateDailyUsageTimeDistribution(data *DailyScheduleEventData, store Store) error {
	return store.calculateDailyUsageTimeDistribution(data.AppId, data.Timestamp)
}
