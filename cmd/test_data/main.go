package main

import (
	"context"
	"fmt"
	"github.com/lt90s/goanalytics/conf"
	"github.com/lt90s/goanalytics/metric/user"
	"github.com/lt90s/goanalytics/storage"
	"github.com/lt90s/goanalytics/storage/mongodb"
	"github.com/lt90s/goanalytics/utils"
	"math/rand"
	"strconv"
)

const (
	// TODO: make it command line argument
	appId = "5cdc1c190419102151e407d5"
)

var (
	end      = utils.TodayTimestamp()
	start    = end - 30*24*60*60
	channels = []string{"huawei", "xiaomi", "appStore", "google"}
	versions = []string{"1.0.0", "1.2.0", "2.0.0"}
)

func setSlotCounterPercent(counter storage.Counter, name string, slots []string) {
	tmpStart := start
	for tmpStart <= end {
		for _, slot := range slots {
			amount := float64(rand.Intn(30)) / 100
			counter.AddSlotCounter(appId, name, slot, tmpStart, float64(amount))
		}
		tmpStart += 24 * 60 * 60
	}
}

func setSlotCounter(counter storage.Counter, name string, slots []string, lower, upper int) {
	tmpStart := start
	for tmpStart <= end {
		for _, slot := range slots {
			amount := rand.Intn(upper-lower) + lower
			counter.AddSlotCounter(appId, name, slot, tmpStart, float64(amount))
		}
		tmpStart += 24 * 60 * 60
	}
}

func setTimeSlotCounter(counter storage.Counter, name string, lower, upper int) {
	tmpStart := start
	for tmpStart <= end {
		for i := 0; i < 24; i++ {
			amount := rand.Intn(upper-lower) + lower
			counter.AddSlotCounter(appId, name, strconv.Itoa(i), tmpStart, float64(amount))
		}
		tmpStart += 24 * 60 * 60
	}
}

func setCPVCounter(counter storage.Counter, name string, lower, upper int) {
	tmpStart := start
	for tmpStart <= end {
		for _, c := range channels {
			for _, p := range []string{"ios", "android"} {
				for _, v := range versions {
					amount := rand.Intn(upper-lower) + lower
					counter.AddSimpleCPVCounter(appId, c, p, v, name, tmpStart, float64(amount))
				}
			}
		}
		tmpStart += 24 * 60 * 60
	}
}

func setSimpleCounterPercent(counter storage.Counter, name string) {
	tmpStart := start
	for tmpStart <= end {
		percent := float64(rand.Intn(100)) / 100.0
		counter.AddSimpleCounter(appId, name, tmpStart, percent)
		tmpStart += 24 * 60 * 60
	}
}

func setSimpleCounter(counter storage.Counter, name string) {
	tmpStart := start
	for tmpStart <= end {
		amount := float64(rand.Intn(100))
		counter.AddSimpleCounter(appId, name, tmpStart, amount)
		tmpStart += 24 * 60 * 60
	}
}

func main() {
	fmt.Println(start, end)
	client := mongodb.DefaultClient
	prefix := conf.GetConfString(conf.MongoDatabasePrefixKey)

	client.Database(prefix + appId).Drop(context.Background())

	counter := mongodb.NewCounter(client, prefix)

	setSimpleCounterPercent(counter, user.DailyActiveNewUserPercentSimpleCounter)

	setTimeSlotCounter(counter, user.OpenAppTimeDistributionSlotCounter, 10, 100)
	setTimeSlotCounter(counter, user.NewUserTimeDistributionSlotCounter, 10, 100)
	setTimeSlotCounter(counter, user.ActiveUserTimeDistributionSlotCounter, 10, 100)

	setCPVCounter(counter, user.DailyActiveCPVCounter, 10, 100)
	setCPVCounter(counter, user.NewUserCPVCounter, 10, 100)
	setCPVCounter(counter, user.OpenAppCPVCounter, 10, 100)

	slots := []string{"1", "2", "3", "4", "5", "6", "7", "15", "30"}
	setSlotCounter(counter, user.NewUserRetentionSlotCounter, slots, 10, 40)
	setSlotCounter(counter, user.ActiveUserRetentionSlotCounter, slots, 100, 500)
	for _, channel := range channels {
		setSlotCounter(counter, user.ChannelNewUserRetentionSlotCounterPrefix+channel, slots, 20, 200)
		setSlotCounter(counter, user.ChannelActiveUserRetentionSlotCounterPrefix+channel, slots, 80, 200)
	}
	setSlotCounter(counter, user.OpenAppCountDistributionSlotCounter, user.OpenAppCountDistributionSlots, 20, 60)
	setSlotCounterPercent(counter, user.DailyActiveUserAffinitySlotCounter, []string{"7", "15", "30"})

	data := storage.CustomizedCounter{
		Name:        "testSimpleEvent",
		DisplayName: "简单事件",
		Type:        "simple",
	}
	counter.AddCustomizedCounter(appId, data)
	setSimpleCounter(counter, data.Name + storage.CustomizedCounterNameSuffix)

	data = storage.CustomizedCounter{
		Name:        "testSlotEvent",
		DisplayName: "分组统计事件",
		Type:        "slot",
		Slots: []string{"分组1", "分组2", "分组3", "分组4"},
	}
	counter.AddCustomizedCounter(appId, data)
	setSlotCounter(counter, data.Name + storage.CustomizedCounterNameSuffix, data.Slots, 10, 100)

	data = storage.CustomizedCounter{
		Name:        "testCPVEvent",
		DisplayName: "分渠道/平台/版本统计事件",
		Type:        "cpv",
		Channels: channels,
		Versions: versions,
	}
	counter.AddCustomizedCounter(appId, data)
	setCPVCounter(counter, data.Name + storage.CustomizedCounterNameSuffix, 10, 100)
}
