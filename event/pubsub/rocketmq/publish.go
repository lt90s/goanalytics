package rocketmq

import (
	rmq "github.com/apache/rocketmq-client-go/core"
	"github.com/lt90s/goanalytics/event/codec"
	log "github.com/sirupsen/logrus"
)

type Publisher struct {
	encoder  codec.DataEncoder
	producer rmq.Producer
	topic    string
}

func NewPublisher(config PublisherConfig, encoder codec.DataEncoder) *Publisher {
	if config.Timeout == 0 {
		config.Timeout = 5
	}
	pConfig := &rmq.ProducerConfig{
		ClientConfig: rmq.ClientConfig{
			GroupID:    config.GroupId,
			NameServer: config.NameServer,
		},
		SendMsgTimeout: config.Timeout,
	}
	producer, err := rmq.NewProducer(pConfig)
	if err != nil {
		panic(err)
	}

	return &Publisher{
		encoder:  encoder,
		producer: producer,
		topic:    config.Topic,
	}
}

func (p *Publisher) Start() error {
	return p.producer.Start()
}

func (p *Publisher) Shutdown() error {
	return p.producer.Shutdown()
}

func (p *Publisher) Publish(event string, data interface{}) error {
	entry := log.WithFields(log.Fields{"event": event, "data": data})
	defer func() {
		if err := recover(); err != nil {
			entry.WithFields(log.Fields{"error": err}).Error("Publish panic!!!")
		}
	}()
	encodeData, err := p.encoder.Encode(data)
	if err != nil {
		entry.Error("Encode data error", "error", err.Error())
		return err
	}
	msg := rmq.Message{
		Topic: p.topic,
		Tags:  event,
		Body:  encodeData,
	}

	entry.WithFields(log.Fields{"msg": msg}).Debug("Publish message")
	err = p.producer.SendMessageOneway(&msg)
	if err != nil {
		entry.WithFields(log.Fields{"error": err.Error()}).Error("Publish error")
	} else {
		entry.Debug("Publish success")
	}
	return err
}
