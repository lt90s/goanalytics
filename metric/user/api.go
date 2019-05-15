package user

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/event/pubsub"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func SetupRoute(iRoute *gin.RouterGroup, oRoute *gin.RouterGroup, publisher pubsub.Publisher, store Store) {
	iGroup := iRoute.Group("/user")
	iGroup.POST("open_app", openAppHandler(publisher))

	//oGroup := oRoute.Group("/user")
}

func openAppHandler(publisher pubsub.Publisher) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		metadata, ok := middlewares.GetMetaData(c)
		if !ok {
			log.Error("[openAppHandler] MetaData missing")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		publisher.Publish(EventUserOpenApp, metadata)
	})
}
