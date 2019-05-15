package storage

type CustomizedCounter struct {
	Name           string   `json:"name" bson:"name"`
	DisplayName    string   `json:"displayName" bson:"displayName"`
	Type           string   `json:"type" bson:"type"`
	Slots          []string `json:"slots" bson:"slots"`
	Channels       []string `json:"channels" bson:"channels"`
	Versions       []string `json:"versions" bson:"versions"`
	YesterdayCount float64  `json:"yesterdayCount" bson:"yesterdayCount"`
	TodayCount     float64  `json:"todayCount" bson:"todayCount"`
}

const CustomizedCounterNameSuffix = "__customized"

func (ce CustomizedCounter) Valid() bool {
	if ce.Name == "" || ce.DisplayName == "" {
		return false
	}

	if ce.Type != "simple" && ce.Type != "slot" && ce.Type != "cpv" {
		return false
	}

	if ce.Type == "slot" && len(ce.Slots) == 0 {
		return false
	}

	if ce.Type == "cpv" && (len(ce.Channels) == 0 || len(ce.Versions) == 0) {
		return false
	}

	return true
}

type Counter interface {
	AddSimpleCounter(appId string, counterName string, dateTimestamp int64, amount float64) error
	GetSimpleCounterSpan(appId string, counterName string, startTimestam, endTimestamp int64) (map[int64]float64, error)
	GetSimpleCounterSum(appId string, counterName string, startTimestam, endTimestamp int64) (float64, error)

	AddSlotCounter(appId string, counterName, slotName string, dateTimestamp int64, amount float64) error
	GetSlotCounterSpan(appId string, target string, start, end int64) (slotCounters SlotCounters, err error)
	GetSlotCounterSum(appId string, target string, start, end int64, slots []string) (sums map[string]float64, err error)

	AddSimpleCPVCounter(appId string, channel, platform, version, counterName string, dateTimestamp int64, amount float64) error
	GetSimpleCPVSumTotal(appId, counterName string, start, end int64) (float64, error)
	GetSimpleCPVSumDate(appId, counterName string, start, end int64) (map[int64]float64, error)
	GetSimpleCPVDateCPV(appId, counterName string, start, end int64) (map[string]map[int64]map[string]float64, error)
	GetSimpleCPVChannelSumDate(appId, counterName, channel string, start, end int64) (map[int64]float64, error)

	AddCustomizedCounter(appId string, data CustomizedCounter) error
	GetCustomizedCounters(appId string) (counters []CustomizedCounter, err error)
	DeleteCustomizedCounter(appId, name, type_ string) error
	GetCustomizedCounter(appId, name, type_ string) (CustomizedCounter, error)
	DropAllCounter(appId string)
}
