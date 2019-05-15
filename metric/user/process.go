package user

import (
	"errors"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/common"
	"github.com/lt90s/goanalytics/event/pubsub"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

func SetupProcessor(subscriber pubsub.Subscriber, store Store) {
	subscriber.Subscribe(EventUserOpenApp, openAppEventHandler(store), middlewares.MetaData{})

	subscriber.Subscribe(DailyScheduleEvent, dailyScheduleEventHandler(store), DailyScheduleEventData{})

	subscriber.Subscribe(common.GlobalEventDropData, dropDataEventHandler(store), common.DropDataRequest{})
}

func dropDataEventHandler(store Store) pubsub.EventHandler {
	return pubsub.EventHandlerFunc(func(data interface{}) error {
		entry := log.WithFields(log.Fields{"data": data, "handler": "dropDataEventHandler"})
		r, ok := data.(*common.DropDataRequest)
		if !ok {
			entry.Warn("data type is not *common.DropDataRequest")
			return errors.New("data type is not *common.DropDataRequest")
		}
		store.dropData(r.AppId)
		return nil
	})
}

func openAppEventHandler(store Store) pubsub.EventHandler {
	return pubsub.EventHandlerFunc(func(data interface{}) error {
		entry := log.WithFields(log.Fields{"data": data, "handler": "openAppEventHandler"})

		metadata, ok := data.(*middlewares.MetaData)
		if !ok {
			entry.Warn("data type is not *middlewares.MetaData")
			return errors.New("data type is not *middlewares.MetaData")
		}

		entry.Debug("Handle open app event")

		store.saveOpenAppData(metadata)

		// open app distribution
		hourSlot := strconv.Itoa(time.Unix(metadata.Timestamp, 0).Hour())
		store.AddSlotCounter(metadata.AppId, OpenAppTimeDistributionSlotCounter, hourSlot, metadata.DateTimestamp, 1.0)

		// open app count
		store.AddSimpleCPVCounter(metadata.AppId, metadata.Channel, metadata.Platform,
			metadata.Version, OpenAppCPVCounter, metadata.DateTimestamp, 1.0)

		// newly registered user
		if store.isUserIdNew(metadata.AppId, metadata.UserId) {
			entry.Debug("Newly registered user")
			store.AddSimpleCPVCounter(metadata.AppId, metadata.Channel, metadata.Platform,
				metadata.Version, NewRegisteredUserCPVCounter, metadata.DateTimestamp, 1.0)
		}

		// new user
		if store.updateUserRecord(metadata) {
			entry.Debug("New user")
			store.AddSlotCounter(metadata.AppId, NewUserTimeDistributionSlotCounter, hourSlot,
				metadata.DateTimestamp, 1.0)
			store.AddSimpleCPVCounter(metadata.AppId, metadata.Channel, metadata.Platform,
				metadata.Version, NewUserCPVCounter, metadata.DateTimestamp, 1.0)
		}

		// FirstOpen update daily active user counter & user retention & active user retention
		if store.deviceFirstOpenToday(metadata) {
			entry.Debug("User first open app today")
			// daily active user
			store.AddSimpleCPVCounter(metadata.AppId, metadata.Channel, metadata.Platform,
				metadata.Version, DailyActiveCPVCounter, metadata.DateTimestamp, 1.0)
			// daily active user hour distribution
			store.AddSlotCounter(metadata.AppId, ActiveUserTimeDistributionSlotCounter,
				hourSlot, metadata.DateTimestamp, 1.0)
			// new user retention
			store.updateNewUserRetention(metadata)
			// active user retention
			store.updateActiveUserRetention(metadata)
			// active user freshness
			store.updateActiveUserFreshness(metadata)
		}
		return nil
	})
}
