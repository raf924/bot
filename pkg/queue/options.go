package queue

type Option interface {
	Apply(q *queue)
}

type queueOptionFunc func(*queue)

func (f queueOptionFunc) Apply(q *queue) {
	f(q)
}

func WithMaxConsumers(max int) Option {
	return queueOptionFunc(func(q *queue) {
		q.maxConsumers = max
	})
}

func WithMaxProducers(max int) Option {
	return queueOptionFunc(func(q *queue) {
		q.maxProducers = max
	})
}
