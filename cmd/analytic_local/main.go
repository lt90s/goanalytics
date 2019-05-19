package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/api/router"
	"github.com/lt90s/goanalytics/conf"
	"github.com/lt90s/goanalytics/event/pubsub/local"
	"github.com/lt90s/goanalytics/metric"
	"github.com/sirupsen/logrus"
)


func main() {
	pubsuber := local.New()
	engine := gin.Default()

	if conf.IsDebug() {
		logrus.SetLevel(logrus.DebugLevel)
	}

	metric.SetupMetricProcessor(pubsuber)
	router.Setup(engine, pubsuber)

	engine.Run(conf.GetConfString(conf.ServerAddr))
}
