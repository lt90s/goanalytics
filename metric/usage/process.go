package usage

import (
	"errors"
	"github.com/lt90s/goanalytics/event/pubsub"
	log "github.com/sirupsen/logrus"
)

func SetupProcessor(subscriber pubsub.Subscriber, store Store) {
	err := subscriber.Subscribe(EventUsageTime, usageTimeEventHandler(store), usageTimeData{})
	if err != nil {
		panic(err)
	}

	err = subscriber.Subscribe(DailyScheduleEvent, dailyScheduleEventHandler(store), DailyScheduleEventData{})
	if err != nil {
		panic(err)
	}
}

func usageTimeEventHandler(store Store) pubsub.EventHandler {
	return pubsub.EventHandlerFunc(func(data interface{}) error {
		entry := log.WithFields(log.Fields{"data": data, "handler": "usageTimeEventHandler"})
		timeData, ok := data.(*usageTimeData)
		if !ok {
			entry.Warn("data type is not type *usageTimeData")
			return errors.New("data type is not *usageTimeData")
		}

		entry.Debug("Handle usage time event")
		err := store.AddSimpleCounter(timeData.MetaData.AppId, UsageSimpleCounter, timeData.MetaData.DateTimestamp, 1.0)
		if err != nil {
			entry.Warn("Add simple Counter UsageSimpleCounter error: ", err.Error())
			return err
		}

		err = store.AddSimpleCounter(timeData.MetaData.AppId, UsageTimeTotalSimpleCounter, timeData.MetaData.DateTimestamp, timeData.Seconds)
		if err != nil {
			if err != nil {
				entry.Warn("Add simple Counter UsageTimeTotalSimpleCounter error: ", err.Error())
				return err
			}
		}


		slot := timeDistribution2Slot(timeData.Seconds)
		err = store.AddSlotCounter(timeData.MetaData.AppId, EachUsageTimeDistributionSlotCounter, slot, timeData.MetaData.DateTimestamp, 1.0)
		if err != nil {
			entry.Warn("add EachUsageTimeDistributionSlotCounter error", "error", err.Error())
			return err
		}

		return nil
	})
}


func timeDistribution2Slot(seconds float64) string {
	if seconds <= 3 {
		return "1-3"
	} else if seconds <= 10 {
		return "4-10"
	} else if seconds <=30 {
		return "11-30"
	} else if seconds <=60 {
		return "31-60"
	} else if seconds <= 180 {
		return "61-180"
	} else if seconds <= 600 {
		return "181-600"
	} else if seconds <= 1800 {
		return "601-1800"
	} else {
		return "1800+"
	}
}