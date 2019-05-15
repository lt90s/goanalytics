package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/storage"
	"github.com/lt90s/goanalytics/storage/mongodb"
	"github.com/lt90s/goanalytics/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
)

type Store interface {
	storage.Counter
	saveOpenAppData(data *middlewares.MetaData) error
	isUserIdNew(appId, userId string) bool
	updateUserRecord(data *middlewares.MetaData) bool
	deviceFirstOpenToday(data *middlewares.MetaData) bool
	updateNewUserRetention(data *middlewares.MetaData)
	updateActiveUserRetention(data *middlewares.MetaData)
	updateActiveUserFreshness(data *middlewares.MetaData)
	getUniqueActiveUserCount(appId string, start, end int64) int
	calcOpenAppCountDistribution(appId string, timestamp int64) error

	dropData(appId string)
}

const (
	openAppDataCollectionName  = "openAppData"
	userCollectionName         = "userCollection"
	deviceActiveCollectionName = "deviceActiveCollection"
)

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

func (ms *mongodbStore) dropData(appId string) {
	ms.client.Database(ms.databasePrefix + appId).Drop(context.Background())
	ms.DropAllCounter(appId)
}

func (ms *mongodbStore) saveOpenAppData(data *middlewares.MetaData) error {
	if data == nil {
		return errors.New("data cannot be nil")
	}
	_, err := ms.database(data.AppId).Collection(openAppDataCollectionName).InsertOne(context.Background(), bson.M{
		"timestamp": data.Timestamp,
		"deviceId":  data.DeviceId,
		"channel":   data.Channel,
		"platform":  data.Platform,
		"version":   data.Version,
		"userId":    data.UserId,
	})
	return err
}

func (ms *mongodbStore) isUserIdNew(appId, userId string) bool {
	filter := bson.M{
		"userId": userId,
	}
	count, err := ms.database(appId).Collection(userCollectionName).CountDocuments(context.Background(), filter)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Warn("isUserIdNew error")
		return false
	}
	return count == 0
}

func (ms *mongodbStore) updateUserRecord(data *middlewares.MetaData) bool {
	filter := bson.M{
		"deviceId": data.DeviceId,
	}
	update := bson.M{
		"$set": bson.M{
			"channel":   data.Channel,
			"platform":  data.Platform,
			"version":   data.Version,
			"userId":    data.UserId,
			"updatedAt": data.Timestamp,
		},
		"$setOnInsert": bson.M{
			"createdAt": data.Timestamp,
		},
	}
	upsert := true
	option := &options.UpdateOptions{
		Upsert: &upsert,
	}
	ctx := context.Background()
	result, err := ms.database(data.AppId).Collection(userCollectionName).UpdateOne(ctx, filter, update, option)
	if err != nil {
		return false
	}
	return result.UpsertedCount > 0
}

func (ms *mongodbStore) deviceFirstOpenToday(data *middlewares.MetaData) bool {
	filter := bson.M{
		"deviceId":  data.DeviceId,
		"timestamp": data.DateTimestamp,
	}
	update := bson.M{
		"$setOnInsert": bson.M{
			"f": 1,
		},
	}
	upsert := true
	option := &options.UpdateOptions{
		Upsert: &upsert,
	}
	result, err := ms.database(data.AppId).Collection(deviceActiveCollectionName).UpdateOne(context.Background(), filter, update, option)
	if err != nil {
		return false
	}
	return result.UpsertedCount > 0
}

func (ms *mongodbStore) getUserCreatedTimestamp(appId, deviceId string) (int64, error) {
	ctx := context.Background()
	filter := bson.M{"deviceId": deviceId}
	option := &options.FindOneOptions{
		Projection: bson.M{"createdAt": 1},
	}
	result := ms.database(appId).Collection(userCollectionName).FindOne(ctx, filter, option)
	var ob struct {
		CreatedAt int64 `bson:"createdAt"`
	}

	if err := result.Decode(&ob); err != nil {
		return 0, err
	}

	return ob.CreatedAt, nil
}

func (ms *mongodbStore) updateNewUserRetention(data *middlewares.MetaData) {
	createdAt, err := ms.getUserCreatedTimestamp(data.AppId, data.DeviceId)
	if err != nil {
		log.Error("[updateNewUserRetention] get user created time error", "error", err.Error())
		return
	}
	createdDateTimestamp := utils.TimestampToDate(createdAt).Unix()
	delta := int((data.Timestamp - createdDateTimestamp) / (24 * 3600))
	if delta > 30 {
		return
	}
	for _, day := range retentionDays {
		if delta == day {
			slot := fmt.Sprintf("%d", day)
			ms.AddSlotCounter(data.AppId, NewUserRetentionSlotCounter, slot, createdDateTimestamp, 1.0)
			ms.AddSlotCounter(data.AppId, ChannelNewUserRetentionSlotCounterPrefix+data.Channel, slot, createdDateTimestamp, 1.0)
			break
		}
	}
}

func (ms *mongodbStore) updateActiveUserRetention(data *middlewares.MetaData) {
	ctx := context.Background()
	filter := bson.M{
		"deviceId": data.DeviceId,
	}

	for _, day := range retentionDays {
		deltaTimestamp := data.DateTimestamp - int64(day*24*3600)
		filter["timestamp"] = deltaTimestamp
		count, err := ms.database(data.AppId).Collection(deviceActiveCollectionName).CountDocuments(ctx, filter)
		if err != nil {
			log.Error("[updateActiveUserRetention] CountDocuments error", "data", data, "day", day)
			continue
		}
		if count > 0 {
			slot := fmt.Sprintf("%d", day)
			ms.AddSlotCounter(data.AppId, ActiveUserRetentionSlotCounter, slot, deltaTimestamp, 1.0)
			ms.AddSlotCounter(data.AppId, ChannelActiveUserRetentionSlotCounterPrefix+data.Channel, slot, deltaTimestamp, 1.0)
		}
	}
}

func (ms *mongodbStore) updateActiveUserFreshness(data *middlewares.MetaData) {
	createdAt, err := ms.getUserCreatedTimestamp(data.AppId, data.DeviceId)
	if err != nil {
		log.Error("[updateNewUserRetention] get user created time error", "error", err.Error())
		return
	}
	createdDateTimestamp := utils.TimestampToDate(createdAt).Unix()
	delta := int((data.Timestamp - createdDateTimestamp) / (24 * 3600))
	if delta > 30 {
		delta = 31
	}
	ms.AddSlotCounter(data.AppId, DailyActiveUserFreshnessSlotCounter, strconv.Itoa(delta), data.DateTimestamp, 1.0)
}

func (ms *mongodbStore) getUniqueActiveUserCount(appId string, start, end int64) int {
	ctx := context.Background()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"timestamp": bson.M{
					"$lte": end,
					"$gte": start,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$deviceId",
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	cursor, err := ms.database(appId).Collection(deviceActiveCollectionName).Aggregate(ctx, pipeline)
	if err != nil {
		return 0
	}
	var tmp struct {
		Count int `bson:"count"`
	}
	if cursor.Next(ctx) {
		cursor.Decode(&tmp)
	}
	return tmp.Count
}

func (ms *mongodbStore) calcOpenAppCountDistribution(appId string, timestamp int64) error {
	end := timestamp + 24*3600
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"timestamp": bson.M{
					"$gte": timestamp,
					"$lt":  end,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$deviceId",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	ctx := context.Background()
	cursor, err := ms.database(appId).Collection(openAppDataCollectionName).Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	var tmp struct {
		Count int `bson:"count"`
	}
	for cursor.Next(ctx) {
		err = cursor.Decode(&tmp)
		if err != nil {
			return err
		}
		var slot string
		count := tmp.Count
		if count <= 2 {
			slot = "1-2"
		} else if count <= 4 {
			slot = "3-4"
		} else if count <= 6 {
			slot = "5-6"
		} else if count <= 8 {
			slot = "7-8"
		} else if count <= 10 {
			slot = "9-10"
		} else if count <= 20 {
			slot = "11-20"
		} else if count <= 30 {
			slot = "21-30"
		} else if count <= 49 {
			slot = "31-49"
		} else {
			slot = "50+"
		}
		ms.AddSlotCounter(appId, OpenAppCountDistributionSlotCounter, slot, timestamp, 1.0)
	}
	return nil
}
