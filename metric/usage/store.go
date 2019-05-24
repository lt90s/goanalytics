package usage

import (
	"context"
	"github.com/lt90s/goanalytics/storage"
	"github.com/lt90s/goanalytics/storage/mongodb"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TODO: setup index

const (
	deviceUsageTimeCollectionName = "deviceUsageTimeCollection"
)

type Store interface {
	storage.Counter
	getTotalUsageTime(appId string, date int64) (float64, error)
	getDeviceCount(appId string, date int64) (int64, error)
	calculateDailyUsageTimeDistribution(appId string, date int64) error
}

type mongodbStore struct {
	storage.Counter
	client         *mongo.Client
	databasePrefix string
}

func NewMongoStore(client *mongo.Client, databasePrefix string) Store {
	return &mongodbStore{
		Counter:        mongodb.NewCounter(client, databasePrefix),
		client:         client,
		databasePrefix: databasePrefix,
	}
}

func (ms *mongodbStore) database(appId string) *mongo.Database {
	return ms.client.Database(ms.databasePrefix + appId)
}

func (ms *mongodbStore) deviceUsageTimeCollection(appId string) *mongo.Collection {
	return ms.database(appId).Collection(deviceUsageTimeCollectionName)
}

func (ms *mongodbStore) addDeviceUsageTime(data *usageTimeData) error {
	ctx := context.Background()
	filter := bson.M{
		"date":     data.MetaData.DateTimestamp,
		"deviceId": data.MetaData.DeviceId,
	}
	update := bson.M{
		"$inc": bson.M{
			"time": data.Seconds,
		},
	}
	upsert := true
	option := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := ms.deviceUsageTimeCollection(data.MetaData.AppId).UpdateOne(ctx, filter, update, &option)
	return err
}

func (ms *mongodbStore) getTotalUsageTime(appId string, date int64) (float64, error) {
	ctx := context.Background()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"date": date,
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"total": bson.M{
					"$sum": "$time",
				},
			},
		},
	}
	cursor, err := ms.deviceUsageTimeCollection(appId).Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	var tmp struct {
		Total float64 `bson:"total"`
	}
	if cursor.Next(ctx) {
		err = cursor.Decode(&tmp)
		if err != nil {
			return 0, err
		}
	}
	return tmp.Total, cursor.Err()
}

func (ms *mongodbStore) getDeviceCount(appId string, date int64) (int64, error) {
	ctx := context.Background()
	filter := bson.M{
		"date": date,
	}
	return ms.deviceUsageTimeCollection(appId).CountDocuments(ctx, filter)
}

func (ms *mongodbStore) calculateDailyUsageTimeDistribution(appId string, date int64) error {
	ctx := context.Background()
	filter := bson.M{
		"date": date,
	}
	option := options.FindOptions{
		Projection: bson.M{
			"time": 1,
		},
	}
	cursor, err := ms.deviceUsageTimeCollection(appId).Find(ctx, filter, &option)
	if err != nil {
		return err
	}
	var tmp struct {
		Time float64 `bson:"time"`
	}
	entry := logrus.WithFields(logrus.Fields{"apppId": appId, "counter": DailyUsageTimeDistributionSlotCounter, "date": date})
	for cursor.Next(ctx) {
		err = cursor.Decode(&tmp)
		if err != nil {
			return err
		}
		slot := timeDistribution2Slot(tmp.Time)
		err = ms.SetSlotCounter(appId, DailyUsageTimeDistributionSlotCounter, slot, date, 1.0)
		if err != nil {
			entry.Warnf("SetSlotCounter error: slot=%v error=%v", slot, err.Error())
		}
	}
	return nil
}
