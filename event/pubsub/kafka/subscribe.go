package kafka

import (
	"context"
	"github.com/lt90s/goanalytics/event/codec"
	"github.com/lt90s/goanalytics/event/pubsub"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"reflect"
	"sync"
	"time"
)

type handlerRecord struct {
	handler  pubsub.EventHandler
	dataType reflect.Type
}

type Subscriber struct {
	decoder       codec.DataDecoder
	reader        *kafka.Reader
	eventMapping  map[string]handlerRecord
	context       context.Context
	cancel        context.CancelFunc
	wg            *sync.WaitGroup
	done          chan struct{}
	processorSema *semaphore.Weighted
}

func NewSubscriber(config kafka.ReaderConfig, decoder codec.DataDecoder, concurrentHandlerCount int64) *Subscriber {
	r := kafka.NewReader(config)
	context, cancel := context.WithCancel(context.Background())
	return &Subscriber{
		decoder:       decoder,
		reader:        r,
		eventMapping:  make(map[string]handlerRecord),
		context:       context,
		cancel:        cancel,
		wg:            &sync.WaitGroup{},
		done:          make(chan struct{}),
		processorSema: semaphore.NewWeighted(concurrentHandlerCount),
	}
}

func (s *Subscriber) Start() {
	go func() {
		for {
			// TODO:
			// `ReadMessage automatically commits offsets when using consumer groups.`
			// If consuming of `msg` fails, it will be lost
			// ignore or retry or else.
			msg, err := s.reader.ReadMessage(s.context)
			if err != nil {
				if err != context.Canceled {
					log.Error("ReadMessage from kafka error, exit... ", "error", err.Error())
				}
				break
			}

			entry := log.WithFields(log.Fields{"event": string(msg.Key)})

			entry.Info("Receive new event")

			record, ok := s.eventMapping[string(msg.Key)]

			if !ok {
				entry.Warn("Event not subscribed")
				continue
			}

			data := reflect.New(record.dataType).Interface()
			err = s.decoder.Decode(string(msg.Value), &data)
			if err != nil {
				entry.Warn("Event data decode error")
				continue
			}

			entry.Debug("Trying to acquire processor")
			err = s.processorSema.Acquire(s.context, 1)
			if err != nil {
				entry.Error("Acquiring processor error ", "error", err.Error())
				continue
			}

			go func() {
				s.wg.Add(1)
				defer func() {
					s.processorSema.Release(1)
					s.wg.Done()
				}()
				// TODO: context
				err = record.handler.Handle(data)
				if err != nil {
					entry.Error("Event handler returns error ", "error", err.Error())
				}
			}()
		}
	}()
}

func (s *Subscriber) Shutdown() {
	go func() {
		time.Sleep(10 * time.Second)
		s.cancel()
	}()
	s.wg.Wait()
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
