package middlewares

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mockAppkeyGetter struct {}

var (
	appId = "testAppId"
	appKey = "testAppKey"
	deviceId = "deviceId"
	channel = "unknown"
	platform = "android"
	version = "0.1.0"
	timestamp = time.Now().Unix()
)

func (m mockAppkeyGetter) GetAppKey(appId string) (string, error) {
	if appId == appId {
		return appKey, nil
	}
	return "", errors.New("appId not exist")
}

func queryString() string {
	qs := fmt.Sprintf("appId=%s&channel=%s&deviceId=%s&platform=%s&timestamp=%d&version=%s",
		appId, channel, deviceId, platform, timestamp, version)
	ss := qs + "&key=" + appKey
	hash := md5.Sum([]byte(ss))
	sign := hex.EncodeToString(hash[:])

	qs += "&sign=" + sign
	return qs
}

func TestMetaDataMiddleware(t *testing.T) {
	middleware := NewMetaDataMiddleware(mockAppkeyGetter{})

	router := gin.Default()
	router.GET("/hello", middleware.Middleware(), func(c *gin.Context) {
		data, ok := GetMetaData(c)
		require.True(t, ok)
		require.Equal(t, data.AppId, appId)
		require.Equal(t, data.DeviceId, deviceId)
		require.Equal(t, data.Channel, channel)
		require.Equal(t, data.Platform, platform)
		require.Equal(t, data.Version, version)
		require.Equal(t, data.Timestamp, timestamp)
		c.Writer.WriteString("hello")
	})

	qs := queryString()
	req := httptest.NewRequest(http.MethodGet, "/hello?" + qs, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "hello", w.Body.String())

	fakeQs := strings.Replace(qs, "android", "ios", 1)
	req = httptest.NewRequest(http.MethodGet, "/hello?" + fakeQs, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}