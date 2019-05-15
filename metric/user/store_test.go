package user

import (
	"context"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/storage/mongodb"
	"github.com/lt90s/goanalytics/utils"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

const (
	appId = "testAppId"
)

func TestMongodbStore_deviceFirstOpenToday(t *testing.T) {
	data := &middlewares.MetaData{
		AppId:     appId,
		DeviceId:  "abc",
		Timestamp: utils.Today().Unix(),
	}
	client := mongodb.DefaultClient
	store := NewMongoStore(client, "test_")
	defer client.Database("test_" + appId).Drop(context.Background())

	flag := store.deviceFirstOpenToday(data)
	require.True(t, flag)

	flag = store.deviceFirstOpenToday(data)
	require.False(t, flag)
}

func TestMongodbStore_isUserIdNew(t *testing.T) {
	data := &middlewares.MetaData{
		AppId:     appId,
		UserId:    "userId",
		Timestamp: utils.Today().Unix(),
	}
	client := mongodb.DefaultClient
	store := NewMongoStore(client, "test_")
	defer client.Database("test_" + appId).Drop(context.Background())

	require.True(t, store.isUserIdNew(data.AppId, data.UserId))
	new := store.updateUserRecord(data)
	require.True(t, new)
	require.False(t, store.isUserIdNew(data.AppId, data.UserId))

}

func TestUpdateNewUserRetention(t *testing.T) {
	data := &middlewares.MetaData{
		AppId:     appId,
		DeviceId:  "abcd",
		Timestamp: utils.TodayDiff(7).Unix(),
	}
	client := mongodb.DefaultClient
	store := NewMongoStore(client, "test_")
	defer client.Database("test_" + appId).Drop(context.Background())

	new := store.updateUserRecord(data)
	require.True(t, new)
	data.DeviceId = "efgh"
	store.updateUserRecord(data)

	for i := 6; i >= 0; i-- {
		data.Timestamp = utils.TodayDiff(i).Unix()
		data.DateTimestamp = data.Timestamp
		if i&1 == 0 {
			data.DeviceId = "abcd"
			data.Channel = "x"
		} else {
			data.DeviceId = "efgh"
			data.Channel = "y"
		}
		store.updateNewUserRetention(data)
	}

	start := utils.TodayDiff(7).Unix()
	end := start
	slotCounters, err := store.GetSlotCounterSpan(data.AppId, NewUserRetentionSlotCounter, start, end)
	require.NoError(t, err)
	slotCounter, ok := slotCounters[start]
	require.True(t, ok)

	for i := 1; i <= 7; i++ {
		count, ok := slotCounter[strconv.Itoa(i)]
		require.True(t, ok)
		require.Equal(t, float64(1), count)
	}

	slotCounters, err = store.GetSlotCounterSpan(data.AppId, ChannelNewUserRetentionSlotCounterPrefix+"x", start, end)
	require.NoError(t, err)
	slotCounter, ok = slotCounters[start]
	for i := 1; i <= 7; i++ {
		count, ok := slotCounter[strconv.Itoa(i)]
		if i&1 == 0 {
			require.False(t, ok)
		} else {
			require.True(t, ok)
			require.Equal(t, float64(1), count)
		}
	}
}

func TestGetActiveUserCount(t *testing.T) {
	client := mongodb.DefaultClient
	store := NewMongoStore(client, "test_")
	//defer client.Database("test_" + appId).Drop(context.Background())

	data := &middlewares.MetaData{
		AppId:         appId,
		DeviceId:      "a",
		DateTimestamp: utils.TodayDiff(1).Unix(),
	}
	store.deviceFirstOpenToday(data)
	data.DeviceId = "b"
	store.deviceFirstOpenToday(data)

	data.DateTimestamp = utils.TodayDiff(0).Unix()
	store.deviceFirstOpenToday(data)
	data.DeviceId = "c"
	store.deviceFirstOpenToday(data)

	count := store.getUniqueActiveUserCount(appId, utils.TodayDiff(1).Unix(), utils.TodayDiff(0).Unix())
	require.Equal(t, 3, count)
}

func BenchmarkUpdateUserRecord(b *testing.B) {
	data := &middlewares.MetaData{
		AppId:     appId,
		UserId:    "userId",
		Timestamp: utils.Today().Unix(),
	}
	client := mongodb.DefaultClient
	store := NewMongoStore(client, "test_")
	defer client.Database("test_" + appId).Drop(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.updateUserRecord(data)
	}
}
