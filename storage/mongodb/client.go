package mongodb

import (
	"context"
	"fmt"
	"github.com/lt90s/goanalytics/conf"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

var DefaultClient *mongo.Client

func init() {
	DefaultClient = NewMongoClient()
}

func NewMongoClient() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(conf.GetConfString(conf.MongoDSNConfKey)))
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
		panic(fmt.Sprintf("mongodb %s not online, err: %s", conf.GetConfString(conf.MongoDSNConfKey), err.Error()))
	}
	return client
}
