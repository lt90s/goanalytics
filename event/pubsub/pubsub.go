package pubsub

type EventHandler interface {
	Handle(data interface{}) error
}

type EventHandlerFunc func(data interface{}) error

func (ehf EventHandlerFunc) Handle(data interface{}) error {
	return ehf(data)
}

type Publisher interface {
	Publish(event string, data interface{}) error
}

type Subscriber interface {
	Subscribe(event string, handler EventHandler, data interface{}) error
}

type PubSuber interface {
	Publisher
	Subscriber
}