package queue

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
)

type Queue interface {
	producer
	consumer
	NewProducer() (*Producer, error)
	NewConsumer() (*Consumer, error)
}

type queueValue struct {
	value     interface{}
	consumers map[string]bool
	producer  string
}

type queue struct {
	maxConsumers int
	rwm          *sync.RWMutex
	c            *sync.Cond
	buffer       []queueValue
	consumers    map[string]struct{}
	vw           ValueWriter
	maxProducers int
	producers    map[string]struct{}
}

func (q *queue) NewProducer() (*Producer, error) {
	if q.maxProducers > 0 && len(q.producers) == q.maxProducers {
		return nil, fmt.Errorf("too many producers")
	}
	id := uuid.NewString()
	q.producers[id] = struct{}{}
	return &Producer{
		id: id,
		q:  q,
	}, nil
}

func (q *queue) NewConsumer() (*Consumer, error) {
	if q.maxConsumers > 0 && len(q.consumers) == q.maxConsumers {
		return nil, fmt.Errorf("too many consumers")
	}
	id := uuid.NewString()
	q.consumers[id] = struct{}{}
	return &Consumer{
		id: id,
		q:  q,
	}, nil
}

func (q *queue) cancel(id string) {
	q.rwm.Lock()
	defer q.rwm.Unlock()
	delete(q.consumers, id)
	for _, qv := range q.buffer {
		delete(qv.consumers, id)
	}
}

func (q *queue) consume(id string, value interface{}) error {
	if _, ok := q.consumers[id]; !ok {
		return fmt.Errorf("not a registered consumer")
	}
	q.rwm.RLock()
	defer func() {
		q.rwm.RUnlock()
	}()
	if len(q.buffer) == 0 {
		q.c.Wait()
	}
	for _, qv := range q.buffer {
		if c, ok := qv.consumers[id]; c || !ok {
			continue
		}
		qv.consumers[id] = true
		err := q.vw.WriteValue(value, qv.value)
		if err != nil {
			return fmt.Errorf("error consuming value: %v", err)
		}
	}
	return nil
}

func (q *queue) produce(id string, value interface{}) error {
	if _, ok := q.producers[id]; !ok {
		return fmt.Errorf("unknown producer")
	}
	q.rwm.Lock()
	toDelete := []int{}
	for i, qv := range q.buffer {
		var allConsumed = true
		for _, consumed := range qv.consumers {
			if !consumed {
				allConsumed = false
				break
			}
		}
		if allConsumed {
			toDelete = append([]int{i}, toDelete...)
		}
	}
	for _, i := range toDelete {
		q.buffer = append(q.buffer[:i], q.buffer[i+1:]...)
	}
	qv := queueValue{
		producer:  id,
		value:     value,
		consumers: map[string]bool{},
	}
	for id := range q.consumers {
		qv.consumers[id] = false
	}
	q.buffer = append(q.buffer, qv)
	q.rwm.Unlock()
	q.c.Signal()
	return nil
}

func newQueue() *queue {
	rwm := &sync.RWMutex{}
	return &queue{
		maxConsumers: 0,
		maxProducers: 0,
		rwm:          rwm,
		vw:           defaultVW,
		c:            sync.NewCond(rwm.RLocker()),
		buffer:       []queueValue{},
		consumers:    map[string]struct{}{},
		producers:    map[string]struct{}{},
	}
}

func NewQueue(options ...Option) Queue {
	q := newQueue()
	for _, option := range options {
		option.Apply(q)
	}
	return q
}
