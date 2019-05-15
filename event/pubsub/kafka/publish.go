package kafka

import (
	"context"
	"github.com/lt90s/goanalytics/event/codec"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Publisher struct {
	encoder codec.DataEncoder
	writer  *kafka.Writer
	context context.Context
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
}

func NewPublisher(config kafka.WriterConfig, encoder codec.DataEncoder) *Publisher {
	w := kafka.NewWriter(config)
	ctx, cancel := context.WithCancel(context.Background())
	return &Publisher{
		encoder: encoder,
		writer:  w,
		context: ctx,
		cancel:  cancel,
		wg:      &sync.WaitGroup{},
	}
}

func (p *Publisher) Start() {}

func (p *Publisher) Shutdown() {
	log.Debug("Kafka publisher shutting down...")
	go func() {
		log.Debug("Wait at most 10 seconds for publishing in process to be done...")
		time.Sleep(10 * time.Second)
		p.cancel()
	}()
	p.wg.Wait()
	log.Debug("Kafka publisher is down now")
}

func (p *Publisher) Publish(event string, data interface{}) error {
	p.wg.Add(1)
	defer p.wg.Done()

	entry := log.WithFields(log.Fields{"event": event, "data": data})
	entry.Debug("Publish message")

	encodedData, err := p.encoder.Encode(data)
	if err != nil {
		entry.Error("Encode data error ", "error", err.Error())
		return err
	}

	err = p.writer.WriteMessages(p.context, kafka.Message{
		Key:   []byte(event),
		Value: []byte(encodedData),
	})
	if err != nil {
		entry.WithFields(log.Fields{"error": err.Error()}).Warn("Publish message failure")
	}
	return err
}
