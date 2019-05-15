package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/api/router"
	"github.com/lt90s/goanalytics/conf"
	"github.com/lt90s/goanalytics/event/pubsub"
	"github.com/lt90s/goanalytics/event/pubsub/local"
	"github.com/lt90s/goanalytics/metric/user"
	"github.com/lt90s/goanalytics/storage/mongodb"
)

func setupProcessor(subscriber pubsub.Subscriber) {
	client := mongodb.DefaultClient
	prefix := conf.GetConfString(conf.MongoDatabasePrefixKey)

	userStore := user.NewMongoStore(client, prefix)
	user.SetupProcessor(subscriber, userStore)
}

func main() {
	pubsuber := local.New()
	engine := gin.Default()

	router.Setup(engine, pubsuber)
	setupProcessor(pubsuber)

	engine.Run(conf.GetConfString(conf.ServerAddr))
}
