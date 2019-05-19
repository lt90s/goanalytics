package usage

import (
	"context"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/storage/mongodb"
	"github.com/lt90s/goanalytics/utils"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
)

const prefix = "metric_usage"
const appId = "test_metric_usage"

var client *mongo.Client
var ms *mongodbStore

func init() {
	client = mongodb.DefaultClient
	ms = NewMongoStore(client, prefix).(*mongodbStore)
}

func drop() {
	client.Database(prefix + appId).Drop(context.Background())
}

func TestAddDeviceUsageTime(t *testing.T) {
	defer drop()

	data := &usageTimeData{
		MetaData: &middlewares.MetaData{
			AppId: appId,
			DeviceId: "a",
			DateTimestamp: utils.TodayTimestamp(),
		},
		Seconds: 12.5,
	}
	err := ms.addDeviceUsageTime(data)
	require.NoError(t, err)

	ctx := context.Background()
	filter := bson.M{"date": data.MetaData.DateTimestamp, "deviceId": data.MetaData.DeviceId}
	sr := ms.deviceUsageTimeCollection(appId).FindOne(ctx, filter)
	require.NoError(t, sr.Err())
	var tmp struct {
		Time float64 `bson:"time"`
	}
	err = sr.Decode(&tmp)
	require.NoError(t, err)
	require.Equal(t, 12.5, tmp.Time)
}

func setupData(t *testing.T) {
	data := &usageTimeData{
		MetaData: &middlewares.MetaData{
			AppId: appId,
			DeviceId: "a",
			DateTimestamp: utils.TodayTimestamp(),
		},
		Seconds: 12.5,
	}
	for i := 0; i < 2; i++ {
		err := ms.addDeviceUsageTime(data)
		require.NoError(t, err)
	}

	data.MetaData.DeviceId = "b"
	data.Seconds = 100.0
	for i := 0; i < 2; i++ {
		err := ms.addDeviceUsageTime(data)
		require.NoError(t, err)
	}
}

func TestGetTotalUsageTime(t *testing.T) {
	defer drop()

	setupData(t)
	total, err := ms.getTotalUsageTime(appId, utils.TodayTimestamp())
	require.NoError(t, err)
	require.Equal(t, 225.0, total)
}

func TestGetDeviceCount(t *testing.T) {
	defer drop()

	setupData(t)

	total, err := ms.getDeviceCount(appId, utils.TodayTimestamp())
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
}

func TestCalculateDailyUsageTimeDistribution(t *testing.T) {
	defer drop()

	setupData(t)
	timestamp := utils.TodayTimestamp()

	err := ms.calculateDailyUsageTimeDistribution(appId, timestamp)
	require.NoError(t, err)

	span, err := ms.GetSlotCounterSpan(appId, DailyUsageTimeDistributionSlotCounter, timestamp, timestamp)
	require.NoError(t, err)

	slotCounter, ok := span[timestamp]
	require.True(t, ok)

	slot := timeDistribution2Slot(25)
	counter, ok := slotCounter[slot]
	require.True(t, ok)
	require.Equal(t, 1.0, counter)

	slot = timeDistribution2Slot(200)
	counter, ok = slotCounter[slot]
	require.True(t, ok)
	require.Equal(t, 1.0, counter)
}