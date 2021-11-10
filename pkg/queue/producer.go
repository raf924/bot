package queue

type Producer interface {
	Produce(value interface{}) error
}

type producer struct {
	id string
	q  produceable
}

func (p *producer) Produce(value interface{}) error {
	return p.q.produce(p.id, value)
}

type produceable interface {
	produce(id string, value interface{}) error
}
