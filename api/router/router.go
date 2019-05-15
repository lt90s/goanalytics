package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/api/authentication"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/conf"
	"github.com/lt90s/goanalytics/event/pubsub"
	"github.com/lt90s/goanalytics/metric"
	"github.com/lt90s/goanalytics/schedule"
	"github.com/lt90s/goanalytics/storage/mongodb"
	"net/http"
)

func Setup(router *gin.Engine, publisher pubsub.Publisher) {
	client := mongodb.DefaultClient
	authStore := authentication.NewMongoStore(client, conf.GetConfString(conf.MongoDatabaseAdminKey))
	counterStore := mongodb.NewCounter(client, conf.GetConfString(conf.MongoDatabasePrefixKey))

	jwtMiddleware := middlewares.NewJwtMiddleware(authStore)
	metadataMiddleware := middlewares.NewMetaDataMiddleware(authStore)

	router.Use(middlewares.ResponseMiddleware)

	authentication.SetupRoute(router, jwtMiddleware, authStore, publisher)

	iRouter := router.Group("/i", metadataMiddleware.Middleware())
	oRouter := router.Group("/o", jwtMiddleware.MiddlewareFunc(), appIdMiddleware)

	InstallCounterEndpoint(iRouter, oRouter, counterStore)

	metric.SetupMetricApi(iRouter, oRouter, publisher)

	schedule.RunScheduler(authStore, publisher)
}

func appIdMiddleware(c *gin.Context) {
	appId := c.Query("appId")
	if appId == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.Set("appId", appId)
	c.Next()
}
