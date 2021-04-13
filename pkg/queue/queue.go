package queue

import (
	"fmt"
	"github.com/google/uuid"
	"log"
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
	next      *queueValue
}

type queue struct {
	maxConsumers int
	locker       sync.Locker
	c            *sync.Cond
	bufferHead   *queueValue
	consumers    map[string]*queueValue
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
	q.consumers[id] = nil
	return &Consumer{
		id: id,
		q:  q,
	}, nil
}

func (q *queue) cancel(id string) {
	q.locker.Lock()
	defer q.locker.Unlock()
	delete(q.consumers, id)
	var current = q.bufferHead
	for current != nil {
		delete(current.consumers, id)
		current = current.next
	}
	q.c.Broadcast()
}

func (q *queue) cleanUp() {
	var previous = (*queueValue)(nil)
	var current = q.bufferHead
	for current != nil {
		var allConsumed = true
		for _, consumed := range current.consumers {
			if !consumed {
				allConsumed = false
				break
			}
		}
		if allConsumed {
			if previous == nil {
				q.bufferHead = current.next
			} else {
				previous.next = current.next
			}
		} else {
			previous = current
		}
		current = current.next
	}
}

func (q *queue) consume(id string) (interface{}, error) {
	q.locker.Lock()
	defer func() {
		q.locker.Unlock()
		r := recover()
		if r != nil {
			log.Panicf("consume panic: %v", r)
		}
	}()
	if _, ok := q.consumers[id]; !ok {
		return nil, fmt.Errorf("not a registered consumer")
	}
	for q.consumers[id] == nil {
		q.c.Wait()
	}
	qv := q.consumers[id]
	qv.consumers[id] = true
	q.consumers[id] = qv.next
	return qv.value, nil
}

func (q *queue) produce(id string, value interface{}) error {
	if _, ok := q.producers[id]; !ok {
		return fmt.Errorf("unknown producer")
	}
	q.locker.Lock()
	defer func() {
		q.c.Broadcast()
	}()
	q.cleanUp()
	qv := &queueValue{
		producer:  id,
		value:     value,
		consumers: map[string]bool{},
	}
	for id := range q.consumers {
		qv.consumers[id] = false
		if q.consumers[id] == nil {
			q.consumers[id] = qv
		}
	}
	q.locker.Unlock()
	if q.bufferHead == nil {
		q.bufferHead = qv
		return nil
	}
	var tail = q.bufferHead
	for ; tail.next != nil; tail = tail.next {
	}
	tail.next = qv
	return nil
}

func newQueue() *queue {
	rwm := &sync.Mutex{}
	return &queue{
		maxConsumers: 0,
		maxProducers: 0,
		locker:       rwm,
		c:            sync.NewCond(rwm),
		bufferHead:   nil,
		consumers:    map[string]*queueValue{},
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
