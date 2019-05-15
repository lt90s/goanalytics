package schedule

import (
	"github.com/lt90s/goanalytics/event/pubsub"
	"github.com/lt90s/goanalytics/metric/user"
	"github.com/lt90s/goanalytics/utils"
	"github.com/whiteshtef/clockwork"
)

type AppIdsGetter interface {
	GetAppIds() []string
}

func RunScheduler(getter AppIdsGetter, publisher pubsub.Publisher) {
	schedule := clockwork.NewScheduler()

	appIds := getter.GetAppIds()

	yesterdayTimestamp := utils.TodayDiff(1).Unix()

	for _, appId := range appIds {
		userDailyData := user.DailyScheduleEventData{
			Timestamp: yesterdayTimestamp,
			AppId:     appId,
		}
		schedule.Schedule().Every().Day().At("1:00").Do(func() {
			publisher.Publish(user.DailyScheduleEvent, &userDailyData)
		})
	}

	go schedule.Run()
}
