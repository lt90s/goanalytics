package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/utils"
	"net/http"
)

func ResponseMiddleware(c *gin.Context) {
	c.Next()
	if c.Writer.Written() {
		return
	}

	value, ok := c.Get("error")
	if ok {
		if err, ok := value.(utils.HttpError); ok {
			c.JSON(err.HttpCode, gin.H{
				"code": err.Code,
				"msg":  err.Msg,
			})
		} else if err, ok := value.(error); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": -1,
				"msg":  err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": -1,
				"msg":  "Unknown error",
			})
		}
		return
	}

	value, ok = c.Get("data")
	if ok {
		c.JSON(http.StatusOK, gin.H{
			"data": value,
		})
	}
}
