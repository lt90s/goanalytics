package user

import (
	"context"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/storage/mongodb"
	"github.com/lt90s/goanalytics/utils"
	"testing"
)

const (
	prefix = "goanalytics_process_test_"
)

// BenchmarkOpenAppEventHandler-12    	     500	   2169938 ns/op
// BenchmarkOpenAppEventHandler-12    	     100	  10152505 ns/op
func BenchmarkOpenAppEventHandler(b *testing.B) {
	client := mongodb.DefaultClient
	store := NewMongoStore(client, prefix)
	defer client.Database(prefix + appId).Drop(context.Background())

	handler := openAppEventHandler(store)
	data := &middlewares.MetaData{
		AppId:         appId,
		DeviceId:      "deviceId",
		Channel:       "channel",
		Platform:      "android",
		Version:       "1.0.0",
		UserId:        "userId",
		Timestamp:     utils.NowTimestamp(),
		DateTimestamp: utils.TodayTimestamp(),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//data.DeviceId += strconv.Itoa(i)
		//data.UserId = strconv.Itoa(i)
		handler.Handle(data)
	}
}
