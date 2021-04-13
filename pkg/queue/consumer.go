package queue

type Consumer struct {
	id string
	q  consumer
}

func (c *Consumer) Consume() (interface{}, error) {
	return c.q.consume(c.id)
}

func (c *Consumer) Cancel() {
	c.q.cancel(c.id)
}

type consumer interface {
	consume(id string) (interface{}, error)
	cancel(id string)
}
