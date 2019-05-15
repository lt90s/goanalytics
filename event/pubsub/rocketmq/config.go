package rocketmq

type PublisherConfig struct {
	Topic      string
	GroupId    string
	NameServer string
	Timeout    int
}

type SubscriberConfig struct {
	Topic      string
	GroupId    string
	NameServer string
}
