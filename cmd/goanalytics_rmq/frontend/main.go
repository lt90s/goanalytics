package main

import (
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/api/router"
	"github.com/lt90s/goanalytics/conf"
	"github.com/lt90s/goanalytics/event/codec"
	"github.com/lt90s/goanalytics/event/pubsub/rocketmq"
	"github.com/sirupsen/logrus"
)

func main() {
	if conf.IsDebug() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	pubConfig := rocketmq.PublisherConfig{
		Topic:      conf.GetConfString(conf.TopicConfKey),
		GroupId:    conf.GetConfString(conf.RMQGroupId),
		NameServer: conf.GetConfString(conf.RMQServeName),
	}
	encoder := codec.NewJsonCodec()
	publisher := rocketmq.NewPublisher(pubConfig, encoder)
	publisher.Start()
	defer publisher.Shutdown()

	engine := gin.Default()

	router.Setup(engine, publisher)

	endless.ListenAndServe(conf.GetConfString(conf.ServerAddr), engine)
}
