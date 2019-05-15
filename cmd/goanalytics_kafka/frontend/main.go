package main

import (
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/api/router"
	"github.com/lt90s/goanalytics/conf"
	"github.com/lt90s/goanalytics/event/codec"
	kafkaPub "github.com/lt90s/goanalytics/event/pubsub/kafka"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func main() {
	if conf.IsDebug() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	pubConfig := kafka.WriterConfig{
		Topic:    conf.GetConfString(conf.TopicConfKey),
		Brokers:  conf.GetConfStringSlice(conf.KafkaBrokersConfKey),
		Balancer: &kafka.RoundRobin{},
	}
	encoder := codec.NewJsonCodec()
	publisher := kafkaPub.NewPublisher(pubConfig, encoder)
	publisher.Start()
	defer publisher.Shutdown()

	engine := gin.Default()

	router.Setup(engine, publisher)

	endless.ListenAndServe(conf.GetConfString(conf.ServerAddr), engine)
}
