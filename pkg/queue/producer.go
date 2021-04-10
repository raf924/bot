package queue

type Producer struct {
	id string
	q  producer
}

func (p *Producer) Produce(value interface{}) error {
	return p.q.produce(p.id, value)
}

type producer interface {
	produce(id string, value interface{}) error
}
