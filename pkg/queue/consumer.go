package queue

type Consumer struct {
	id string
	q  consumer
}

func (c *Consumer) Consume(value interface{}) error {
	return c.q.consume(c.id, value)
}

func (c *Consumer) Cancel() {
	c.q.cancel(c.id)
}

type consumer interface {
	consume(id string, value interface{}) error
	cancel(id string)
}
