package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/storage/mongodb"
	"github.com/lt90s/goanalytics/utils"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	appId          = "testAppId"
	databasePrefix = "testCounter_"
)

func TestCounter_SimpleCounter(t *testing.T) {
	client := mongodb.DefaultClient
	mongoCounter := mongodb.NewCounter(client, databasePrefix)
	defer client.Database(databasePrefix + appId).Drop(context.Background())

	timestamp := utils.TodayTimestamp()
	mongoCounter.AddSimpleCounter(appId, "foo", timestamp, 1.4)

	router := gin.Default()
	router.Use(appIdMiddleware, middlewares.ResponseMiddleware)
	InstallCounterEndpoint(router.Group("/test"), mongoCounter)

	data := counterDescriptorData{
		Descriptors: []counterDescriptor{
			{
				Type:     "simple",
				Name:     "foo",
				Operator: "sum",
				Start:    timestamp,
				End:      timestamp,
			},
		},
	}
	// sum operator
	s, err := json.Marshal(data)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/test/counter?appId="+appId, bytes.NewBuffer(s))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, `{"data":{"foo":1.4}}`, w.Body.String())

	// span operator
	data.Descriptors[0].Operator = "span"
	s, err = json.Marshal(data)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodPost, "/test/counter?appId="+appId, bytes.NewBuffer(s))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	cmp := fmt.Sprintf(`{"data":{"foo":{"%d":1.4}}}`, timestamp)
	require.Equal(t, cmp, w.Body.String())

	// unsupported operator
	data.Descriptors[0].Operator = "bar"
	s, err = json.Marshal(data)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodPost, "/test/counter?appId="+appId, bytes.NewBuffer(s))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCounter_SlotCounter(t *testing.T) {
	client := mongodb.DefaultClient
	mongoCounter := mongodb.NewCounter(client, databasePrefix)
	defer client.Database(databasePrefix + appId).Drop(context.Background())

	timestamp := utils.TodayTimestamp()
	mongoCounter.AddSlotCounter(appId, "foo", "fooSlot", timestamp, 1.4)
	mongoCounter.AddSlotCounter(appId, "foo", "barSlot", timestamp, 2.8)
	mongoCounter.AddSlotCounter(appId, "foo", "bazSlot", timestamp, 1.0)

	router := gin.Default()
	router.Use(appIdMiddleware, middlewares.ResponseMiddleware)
	InstallCounterEndpoint(router.Group("/test"), mongoCounter)

	data := counterDescriptorData{
		Descriptors: []counterDescriptor{
			{
				Type:     "slot",
				Name:     "foo",
				Operator: "span",
				Start:    timestamp,
				End:      timestamp,
			},
		},
	}

	s, err := json.Marshal(data)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/test/counter?appId="+appId, bytes.NewBuffer(s))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	t.Log(w.Body.String())
	cmp := fmt.Sprintf(`{"data":{"foo":{"%d":{"barSlot":2.8,"bazSlot":1,"fooSlot":1.4}}}}`, timestamp)
	require.Equal(t, cmp, w.Body.String())
}
