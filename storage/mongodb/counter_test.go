package mongodb

import (
	"context"
	"fmt"
	"github.com/lt90s/goanalytics/utils"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strconv"
	"testing"
	"time"
)

const (
	mongoDBUri = "mongodb://127.0.0.1:27017"
	appId      = "testAppId"
)

func newMongoClient() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoDBUri))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		panic(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(fmt.Sprintf("mongodb %s not online, err: %s", mongoDBUri, err.Error()))
	}
	return client
}

func TestCounter_AddSimpleCounter_GetSimpleCounterSpan(t *testing.T) {
	mongoCounter := NewCounter(newMongoClient(), "goanalytics").(*counter)
	defer mongoCounter.database(appId).Drop(context.Background())

	timestamp := utils.TodayTimestamp()
	err := mongoCounter.AddSimpleCounter(appId, "foo", timestamp, 2.4)
	require.NoError(t, err)

	err = mongoCounter.AddSimpleCounter(appId, "foo", timestamp, 5.2)
	require.NoError(t, err)

	span, err := mongoCounter.GetSimpleCounterSpan(appId, "foo", timestamp, timestamp)
	require.NoError(t, err)
	require.Equal(t, 1, len(span))

	count, ok := span[timestamp]
	require.True(t, ok)
	require.Equal(t, 7.6, count)
}

func TestCounter_SetSimpleCounter_GetSimpleCounterSpan(t *testing.T) {
	mongoCounter := NewCounter(newMongoClient(), "goanalytics").(*counter)
	defer mongoCounter.database(appId).Drop(context.Background())

	timestamp := utils.TodayTimestamp()
	err := mongoCounter.SetSimpleCounter(appId, "foo", timestamp, 2.4)
	require.NoError(t, err)

	err = mongoCounter.SetSimpleCounter(appId, "foo", timestamp, 5.2)
	require.NoError(t, err)

	span, err := mongoCounter.GetSimpleCounterSpan(appId, "foo", timestamp, timestamp)
	require.NoError(t, err)
	require.Equal(t, 1, len(span))

	count, ok := span[timestamp]
	require.True(t, ok)
	require.Equal(t, 5.2, count)
}

func TestCounter_GetSimpleCounterSum(t *testing.T) {
	mongoCounter := NewCounter(newMongoClient(), "goanalytics").(*counter)
	defer mongoCounter.database(appId).Drop(context.Background())

	err := mongoCounter.AddSimpleCounter(appId, "foo", utils.TodayTimestamp(), 2.4)
	require.NoError(t, err)

	err = mongoCounter.AddSimpleCounter(appId, "foo", utils.TodayDiff(1).Unix(), 3.2)
	require.NoError(t, err)

	err = mongoCounter.AddSimpleCounter(appId, "foo", utils.TodayDiff(2).Unix(), 4.8)
	require.NoError(t, err)

	sum, err := mongoCounter.GetSimpleCounterSum(appId, "foo", utils.TodayDiff(2).Unix(), utils.TodayTimestamp())
	require.NoError(t, err)
	require.Equal(t, 10.4, sum)
}

func TestCounter_AddSlotCounter_GetSlotCounterSpan(t *testing.T) {
	mongoCounter := NewCounter(newMongoClient(), "goanalytics").(*counter)
	defer mongoCounter.database(appId).Drop(context.Background())

	err := mongoCounter.AddSlotCounter(appId, "foo", "bar", utils.TodayTimestamp(), 1)
	require.NoError(t, err)

	err = mongoCounter.AddSlotCounter(appId, "foo", "baz", utils.TodayTimestamp(), 2.4)
	require.NoError(t, err)

	span, err := mongoCounter.GetSlotCounterSpan(appId, "foo", utils.TodayTimestamp(), utils.TodayTimestamp())
	require.NoError(t, err)
	require.Len(t, span, 1)

	slotCounter, ok := span[utils.TodayTimestamp()]
	require.True(t, ok)

	bar, ok := slotCounter["bar"]
	require.True(t, ok)
	require.Equal(t, 1.0, bar)

	baz, ok := slotCounter["baz"]
	require.True(t, ok)
	require.Equal(t, 2.4, baz)
}

func TestCounter_GetSlotCounterPartialSlotSum(t *testing.T) {
	mongoCounter := NewCounter(newMongoClient(), "goanalytics").(*counter)
	defer mongoCounter.database(appId).Drop(context.Background())

	for i := 1; i <= 24; i++ {
		err := mongoCounter.AddSlotCounter(appId, "foo", strconv.Itoa(i), utils.TodayTimestamp(), 1.0)
		require.NoError(t, err)
	}

	slots := []string{"1", "2", "3", "4", "5", "6", "7"}
	sum := mongoCounter.GetSlotCounterPartialSlotSum(appId, "foo", utils.TodayTimestamp(), slots)
	require.Equal(t, 7.0, sum)
}

func TestCounter_GetSlotCounterSum(t *testing.T) {
	mongoCounter := NewCounter(newMongoClient(), "goanalytics").(*counter)
	defer mongoCounter.database(appId).Drop(context.Background())

	for i := 0; i < 7; i++ {
		err := mongoCounter.AddSlotCounter(appId, "foo", "bar", utils.TodayDiff(i).Unix(), 1.2)
		require.NoError(t, err)

		err = mongoCounter.AddSlotCounter(appId, "foo", "baz", utils.TodayDiff(i).Unix(), 2.4)
		require.NoError(t, err)
	}

	sums, err := mongoCounter.GetSlotCounterSum(appId, "foo", utils.TodayDiff(6).Unix(), utils.TodayTimestamp(), []string{"bar", "baz"})
	require.NoError(t, err)
	require.Len(t, sums, 2)

	barSum, ok := sums["bar"]
	require.True(t, ok)
	require.Equal(t, 1.2*7, barSum)

	bazSum, ok := sums["baz"]
	require.True(t, ok)
	require.Equal(t, 2.4*7, bazSum)
}

func TestCounter_SetSlotCounter_GetSlotCounterSpan(t *testing.T) {
	mongoCounter := NewCounter(newMongoClient(), "goanalytics").(*counter)
	defer mongoCounter.database(appId).Drop(context.Background())

	err := mongoCounter.SetSlotCounter(appId, "foo", "bar", utils.TodayTimestamp(), 1)
	require.NoError(t, err)

	err = mongoCounter.SetSlotCounter(appId, "foo", "bar", utils.TodayTimestamp(), 1)
	require.NoError(t, err)

	err = mongoCounter.SetSlotCounter(appId, "foo", "baz", utils.TodayTimestamp(), 2.4)
	require.NoError(t, err)

	err = mongoCounter.SetSlotCounter(appId, "foo", "baz", utils.TodayTimestamp(), 2.4)
	require.NoError(t, err)

	span, err := mongoCounter.GetSlotCounterSpan(appId, "foo", utils.TodayTimestamp(), utils.TodayTimestamp())
	require.NoError(t, err)
	require.Len(t, span, 1)

	slotCounter, ok := span[utils.TodayTimestamp()]
	require.True(t, ok)

	bar, ok := slotCounter["bar"]
	require.True(t, ok)
	require.Equal(t, 1.0, bar)

	baz, ok := slotCounter["baz"]
	require.True(t, ok)
	require.Equal(t, 2.4, baz)
}

func TestCounter_GetSimpleCPVSumTotal(t *testing.T) {
	mongoCounter := NewCounter(newMongoClient(), "goanalytics").(*counter)
	defer mongoCounter.database(appId).Drop(context.Background())

	for i := 0; i < 10; i++ {
		err := mongoCounter.AddSimpleCPVCounter(appId, "c0", "p0", "v0", "cpv", utils.TodayTimestamp(), 1.0)
		require.NoError(t, err)
	}

	sum, err := mongoCounter.GetSimpleCPVSumTotal(appId, "cpv", utils.TodayTimestamp(), utils.TodayTimestamp())
	require.NoError(t, err)
	require.Equal(t, 10.0, sum)
}

func TestCounter_GetSimpleCPVSumDate(t *testing.T) {
	mongoCounter := NewCounter(newMongoClient(), "test_").(*counter)
	defer mongoCounter.database(appId).Drop(context.Background())

	for i := 0; i < 10; i++ {
		err := mongoCounter.AddSimpleCPVCounter(appId, "c0", "p0", "v0", "cpv", utils.TodayTimestamp(), 1.0)
		require.NoError(t, err)
		err = mongoCounter.AddSimpleCPVCounter(appId, "c0", "p0", "v0", "cpv", utils.TodayDiff(1).Unix(), 2.0)
		require.NoError(t, err)
	}

	sums, err := mongoCounter.GetSimpleCPVSumDate(appId, "cpv", utils.TodayDiff(1).Unix(), utils.TodayTimestamp())
	require.NoError(t, err)
	t.Log(sums)
	require.Len(t, sums, 2)
	sum, ok := sums[utils.TodayDiff(1).Unix()]
	require.True(t, ok)
	require.Equal(t, 20.0, sum)

	sum, ok = sums[utils.TodayTimestamp()]
	require.True(t, ok)
	require.Equal(t, 10.0, sum)
}

func TestCounter_SetSimpleCPVCounter(t *testing.T) {
	mongoCounter := NewCounter(newMongoClient(), "goanalytics").(*counter)
	defer mongoCounter.database(appId).Drop(context.Background())

	err := mongoCounter.SetSimpleCPVCounter(appId, "c0", "p0", "v0", "cpv", utils.TodayTimestamp(), 1.0)
	require.NoError(t, err)

	err = mongoCounter.SetSimpleCPVCounter(appId, "c0", "p0", "v0", "cpv", utils.TodayTimestamp(), 2.0)
	require.NoError(t, err)

	sum, err := mongoCounter.GetSimpleCPVSumTotal(appId, "cpv", utils.TodayTimestamp(), utils.TodayTimestamp())
	require.NoError(t, err)
	require.Equal(t, 2.0, sum)
}
