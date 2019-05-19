package usage

import "github.com/lt90s/goanalytics/api/middlewares"

const (
	EventUsageTime = "EventUsageTime"
)

const (
	UsageSimpleCounter                    = "UsageSimpleCounter"
	UsageTimeTotalSimpleCounter           = "UsageTimeTotalSimpleCounter"
	EachUsageTimeDistributionSlotCounter  = "EachUsageTimeDistributionSlotCounter"
	EachUsageAverageTimeSimpleCounter     = "EachUsageAverageTimeSimpleCounter"
	DailyUsageTimeDistributionSlotCounter = "DailyUsageTimeDistributionSlotCounter"
	DailyUsageAverageTimeSimpleCountet    = "DailyUsageAverageTimeSimpleCountet"
)

const (
	DailyScheduleEvent = "UsageDailyScheduleEvent"
)

var (
	TimeDistributionSlotsMapping = map[string]string {
		"1-3": "1秒-3秒",
		"3-10": "4秒-10秒",
		"11-30": "11秒-30秒",
		"31-60": "31秒-1分",
		"61-180": "1分-3分",
		"181-600": "3分-10分",
		"601-1800": "10分-30分",
		"1800+": "30分+",
	}
)

type usageTimeRequestData struct {
	Seconds float64 `json:"seconds"'`
}

type usageTimeData struct {
	MetaData *middlewares.MetaData `json:"metadata"`
	Seconds  float64               `json:"seconds"`
}