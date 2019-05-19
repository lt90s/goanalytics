package usage

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/event/pubsub"
	"github.com/lt90s/goanalytics/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func SetupRoute(iRoute *gin.RouterGroup, oRoute *gin.RouterGroup, publisher pubsub.Publisher, store Store) {
	iGroup := iRoute.Group("/usage")
	iGroup.POST("/time", usageTimeHandler(publisher))
}

func usageTimeHandler(publisher pubsub.Publisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestData usageTimeRequestData
		err := c.ShouldBindJSON(&requestData)
		if err != nil {
			c.Set("error", utils.ParamError)
			return
		}

		if requestData.Seconds < 0.1 {
			return
		}

		metadata, ok := middlewares.GetMetaData(c)
		if !ok {
			log.Error("[openAppHandler] MetaData missing")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		data := &usageTimeData{
			MetaData: metadata,
			Seconds:  requestData.Seconds,
		}

		publisher.Publish(EventUsageTime, data)
	}
}
