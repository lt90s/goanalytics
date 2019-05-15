package rocketmq

import (
	"errors"
	"github.com/lt90s/goanalytics/event/codec"
	"github.com/lt90s/goanalytics/event/pubsub"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type fooData struct {
	Foo int    `json:"foo"`
	Bar string `json:"bar"`
}

func TestPublisher_Publish(t *testing.T) {
	pubConfig := PublisherConfig{
		Topic:      "testTopicA",
		GroupId:    "testGroup",
		NameServer: "192.168.254.250:9876",
	}

	codecs := codec.NewJsonCodec()

	publisher := NewPublisher(pubConfig, codecs)
	publisher.Start()
	defer publisher.Shutdown()

	err := publisher.Publish("foo", fooData{1, "bar"})
	require.NoError(t, err)
}

func TestSubscriber_Subscribe(t *testing.T) {
	subConfig := SubscriberConfig{
		Topic:      "testTopicA",
		GroupId:    "testConsumer",
		NameServer: "192.168.254.250:9876",
	}
	codecs := codec.NewJsonCodec()
	subscriber := NewSubscriber(subConfig, codecs)

	exit := make(chan struct{})
	subscriber.Subscribe("foo", pubsub.EventHandlerFunc(func(data interface{}) error {
		fb, ok := data.(*fooData)
		require.True(t, ok)
		require.Equal(t, 1, fb.Foo)
		require.Equal(t, "bar", fb.Bar)
		select {
		case exit <- struct{}{}:
		default:
			return nil
		}
		return nil
	}), fooData{})

	subscriber.Start()
	defer subscriber.Shutdown()

	timer := time.NewTimer(10 * time.Second)
	select {
	case <-exit:
	case <-timer.C:
		require.NoError(t, errors.New("timeout"))
	}
}
