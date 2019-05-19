package main

import (
	"github.com/lt90s/goanalytics/conf"
	"github.com/lt90s/goanalytics/event/codec"
	"github.com/lt90s/goanalytics/event/pubsub/rocketmq"
	"github.com/lt90s/goanalytics/metric"
	"os"
	"os/signal"
	"syscall"
)


func main() {
	subConfig := rocketmq.SubscriberConfig{
		Topic:      conf.GetConfString(conf.TopicConfKey),
		GroupId:    conf.GetConfString(conf.RMQGroupId),
		NameServer: conf.GetConfString(conf.RMQServeName),
	}
	decoder := codec.NewJsonCodec()
	subscriber := rocketmq.NewSubscriber(subConfig, decoder)
	subscriber.Start()
	defer subscriber.Shutdown()

	metric.SetupMetricProcessor(subscriber)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	<-sigChan
}
