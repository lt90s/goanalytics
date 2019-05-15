package middlewares

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/conf"
	"github.com/lt90s/goanalytics/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

const metaDataKey = "_metadata"

type AppKeyGetter interface {
	GetAppKey(appId string) (string, error)
}

type MetaData struct {
	AppId         string
	DeviceId      string
	Channel       string
	Platform      string
	Version       string
	UserId        string
	Timestamp     int64
	DateTimestamp int64
}

type MetaDataMiddleware struct {
	appKeyGetter AppKeyGetter
}

func NewMetaDataMiddleware(appKeyGetter AppKeyGetter) MetaDataMiddleware {
	return MetaDataMiddleware{appKeyGetter}
}

func (m MetaDataMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		timestampS := c.Query("timestamp")
		timestamp, err := strconv.ParseInt(timestampS, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		data := &MetaData{
			AppId:     c.Query("appId"),
			DeviceId:  c.Query("deviceId"),
			Channel:   c.Query("channel"),
			Platform:  c.Query("platform"),
			Version:   c.Query("version"),
			UserId:    c.Query("userId"),
			Timestamp: timestamp,
		}
		if !m.validateMetaData(data, c.Query("sign")) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.Set(metaDataKey, data)
		c.Next()
	}
}

func (m MetaDataMiddleware) validateMetaData(data *MetaData, sign string) bool {
	logEntry := log.WithFields(log.Fields{"metadata": data})
	// do not check sign when debug
	if !conf.IsDebug() {
		key, err := m.appKeyGetter.GetAppKey(data.AppId)
		if err != nil {
			logEntry.Debug("GetAppKey failed: ", err.Error())
			return false
		}

		s := fmt.Sprintf("appId=%s&channel=%s&deviceId=%s&platform=%s&timestamp=%d&version=%s&key=%s",
			data.AppId, data.Channel, data.DeviceId, data.Platform, data.Timestamp, data.Version, key)
		hash := md5.Sum([]byte(s))
		if sign != hex.EncodeToString(hash[:]) {
			logEntry.Debug("sign mismatch")
			return false
		}
	}

	if data.AppId == "" {
		return false
	}

	if data.DeviceId == "" {
		return false
	}

	if data.Channel == "" {
		return false
	}

	platform := strings.ToLower(data.Platform)
	if platform != "ios" && platform != "android" {
		return false
	}
	data.Platform = platform

	if data.Version == "" {
		return false
	}

	data.DateTimestamp = utils.TimestampToDate(data.Timestamp).Unix()
	return true
}

func GetMetaData(c *gin.Context) (*MetaData, bool) {
	value, ok := c.Get("_metadata")
	if !ok {
		return nil, false
	}

	data, ok := value.(*MetaData)
	if !ok {
		return nil, false
	}
	return data, true
}
