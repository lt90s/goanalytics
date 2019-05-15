package local

import (
	"github.com/lt90s/goanalytics/event/pubsub"
	log "github.com/sirupsen/logrus"
)

type pubSubLocal struct {
	eventHandlerMapping map[string]pubsub.EventHandler
}

func New() pubsub.PubSuber {
	return &pubSubLocal{
		eventHandlerMapping: make(map[string]pubsub.EventHandler),
	}
}

func (psl *pubSubLocal) Publish(event string, data interface{}) error {
	entry := log.WithFields(log.Fields{"event": event, "data": data})
	entry.Debug("New event")

	handler, ok := psl.eventHandlerMapping[event]
	if !ok {
		entry.Warn("Event not subscribed")
	}
	if handler != nil {
		err := handler.Handle(data)
		if err != nil {
			entry.WithFields(log.Fields{"error": err.Error()}).Warn("Event handler error")
		}
	}
	return nil
}

func (psl *pubSubLocal) Subscribe(event string, handler pubsub.EventHandler, data interface{}) error {
	entry := log.WithFields(log.Fields{"event": event})
	if _, ok := psl.eventHandlerMapping[event]; ok {
		entry.Panic("Event already subscribed")
	}
	psl.eventHandlerMapping[event] = handler
	entry.Info("Event subscribe success")
	return nil
}
