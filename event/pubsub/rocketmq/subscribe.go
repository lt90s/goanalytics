package rocketmq

import (
	rmq "github.com/apache/rocketmq-client-go/core"
	"github.com/lt90s/goanalytics/event/codec"
	"github.com/lt90s/goanalytics/event/pubsub"
	log "github.com/sirupsen/logrus"
	"reflect"
)

type handlerRecord struct {
	handler  pubsub.EventHandler
	dataType reflect.Type
}

type Subscriber struct {
	decoder      codec.DataDecoder
	consumer     rmq.PushConsumer
	topic        string
	eventMapping map[string]handlerRecord
}

func NewSubscriber(config SubscriberConfig, decoder codec.DataDecoder) *Subscriber {
	pConfig := &rmq.PushConsumerConfig{
		ClientConfig: rmq.ClientConfig{
			GroupID:    config.GroupId,
			NameServer: config.NameServer,
		},
		Model: rmq.Clustering,
	}
	consumer, err := rmq.NewPushConsumer(pConfig)
	if err != nil {
		panic(err)
	}

	return &Subscriber{
		decoder:      decoder,
		consumer:     consumer,
		topic:        config.Topic,
		eventMapping: make(map[string]handlerRecord),
	}
}

func (s *Subscriber) Start() {
	s.consumer.Subscribe(s.topic, "*", func(msg *rmq.MessageExt) rmq.ConsumeStatus {
		entry := log.WithFields(log.Fields{"event": msg.Tags, "data": msg.Body})
		entry.Info("New event received")

		event := msg.Tags
		record, ok := s.eventMapping[event]
		if !ok {
			entry.Warn("Event not subscribed")
			return rmq.ConsumeSuccess
		}

		data := reflect.New(record.dataType).Interface()
		err := s.decoder.Decode(msg.Body, &data)

		if err != nil {
			entry.Error("Decode data error", "error", err.Error())
			return rmq.ConsumeSuccess
		}
		record.handler.Handle(data)
		return rmq.ConsumeSuccess
	})

	err := s.consumer.Start()
	if err != nil {
		panic(err)
	}
}

func (s *Subscriber) Shutdown() error {
	return s.consumer.Shutdown()
}

func (s *Subscriber) Subscribe(event string, handler pubsub.EventHandler, data interface{}) error {
	entry := log.WithFields(log.Fields{"event": event})
	if _, ok := s.eventMapping[event]; ok {
		entry.Panic("Event already subscribed")
	}

	if handler == nil {
		entry.Panic("Event handler must not be nil")
	}

	s.eventMapping[event] = handlerRecord{
		handler:  handler,
		dataType: reflect.TypeOf(data),
	}
	entry.Info("Event subscribe success")
	return nil
}
