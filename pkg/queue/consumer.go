package queue

type Consumer interface {
	Consume() (interface{}, error)
	Cancel()
}

type consumer struct {
	id string
	q  consumable
}

var _ Consumer = (*consumer)(nil)

func (c *consumer) Consume() (interface{}, error) {
	return c.q.consume(c.id)
}

func (c *consumer) Cancel() {
	c.q.cancel(c.id)
}

type consumable interface {
	consume(id string) (interface{}, error)
	cancel(id string)
}
