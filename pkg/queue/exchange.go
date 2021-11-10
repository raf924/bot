package queue

type Exchange interface {
	Producer
	Consumer
}

type exchange struct {
	Producer
	Consumer
}

func NewExchange(producerQueue, consumerQueue Queue) (Exchange, error) {
	producer, err := producerQueue.NewProducer()
	if err != nil {
		return nil, err
	}
	consumer, err := consumerQueue.NewConsumer()
	if err != nil {
		return nil, err
	}
	return &exchange{
		Producer: producer,
		Consumer: consumer,
	}, nil
}
