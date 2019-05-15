package local

import (
	"github.com/lt90s/goanalytics/event/pubsub"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPubSubLocal_Subscribe(t *testing.T) {
	ps := New()
	err := ps.Subscribe("foo", nil, nil)
	require.NoError(t, err)

	require.Panics(t, func() {
		ps.Subscribe("foo", nil, nil)
	})
}

func TestPubSubLocal_Publish(t *testing.T) {
	ps := New()
	var foo string
	err := ps.Subscribe("foo", pubsub.EventHandlerFunc(func(data interface{}) error {
		s, ok := data.(string)
		require.True(t, ok)
		foo = s
		return nil
	}), nil)
	require.NoError(t, err)

	ps.Publish("foo", "foo")
	require.Equal(t, "foo", foo)
}
