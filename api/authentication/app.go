package authentication

import (
	"context"
	"github.com/lt90s/goanalytics/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	appKeyLength  = 64
	appCollection = "applicationCollection"
)

func (ms *mongoStore) appCollection() *mongo.Collection {
	return ms.client.Database(ms.database).Collection(appCollection)
}

func (ms *mongoStore) GetAppIds() []string {
	ctx := context.Background()
	option := &options.FindOptions{
		Projection: bson.M{"_id": 1},
	}
	cursor, err := ms.appCollection().Find(ctx, bson.M{}, option)
	if err != nil {
		return nil
	}
	appIds := make([]string, 0, 10)
	var info AppInfo
	for cursor.Next(ctx) {
		err = cursor.Decode(&info)
		if err != nil {
			return nil
		}
		appIds = append(appIds, info.MongoId.Hex())
	}
	return appIds
}

func (ms *mongoStore) GetApps() (infos []AppInfo, err error) {
	ctx := context.Background()
	cursor, err := ms.appCollection().Find(ctx, bson.M{})
	if err != nil {
		return
	}
	var info AppInfo
	for cursor.Next(ctx) {
		err = cursor.Decode(&info)
		if err != nil {
			return
		}
		info.AppId = info.MongoId.Hex()
		infos = append(infos, info)
	}
	return
}

func (ms *mongoStore) CreateApp(name, description string) (info AppInfo, err error) {
	appKey, err := utils.RandomHexStringKey(appKeyLength)
	if err != nil {
		return
	}
	ctx := context.Background()
	now := utils.NowTimestamp()
	result, err := ms.appCollection().InsertOne(ctx, bson.M{
		"name":        name,
		"description": description,
		"appKey":      appKey,
		"createdAt":   now,
	})

	if err != nil {
		return
	}
	info = AppInfo{
		Name:        name,
		AppId:       result.InsertedID.(primitive.ObjectID).Hex(),
		AppKey:      appKey,
		Description: description,
		CreatedAt:   now,
	}
	return
}

func (ms *mongoStore) GetAppKey(appId string) (key string, err error) {
	ctx := context.Background()
	id, err := primitive.ObjectIDFromHex(appId)
	if err != nil {
		return
	}
	filter := bson.M{"_id": id}
	option := &options.FindOneOptions{
		Projection: bson.M{"appKey": 1},
	}
	result := ms.appCollection().FindOne(ctx, filter, option)

	if err = result.Err(); err != nil {
		return
	}

	var info AppInfo
	if err = result.Decode(&info); err != nil {
		return
	}

	key = info.AppKey

	return
}

func (ms *mongoStore) DeleteApp(appId string) error {
	ctx := context.Background()
	id, err := primitive.ObjectIDFromHex(appId)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": id}

	_, err = ms.appCollection().DeleteOne(ctx, filter)

	return err
}
