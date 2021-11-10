package queue

import (
	"fmt"
	"github.com/segmentio/ksuid"
	"sync"
)

type Queue interface {
	produceable
	consumable
	NewProducer() (Producer, error)
	NewConsumer() (Consumer, error)
}

type linkedBuffer struct {
	root *bufferValue
	rwm  *sync.RWMutex
	c    *sync.Cond
}

type bufferValue struct {
	value      interface{}
	next       *bufferValue
	producerId string
}

type queue struct {
	rLocker         sync.Locker
	wLocker         sync.Locker
	consumerBuffers map[string]*linkedBuffer
}

func newQueue() *queue {
	rwm := &sync.RWMutex{}
	return &queue{
		rLocker:         rwm.RLocker(),
		wLocker:         rwm,
		consumerBuffers: map[string]*linkedBuffer{},
	}
}

func (q *queue) produce(id string, value interface{}) error {
	q.rLocker.Lock()
	for _, buffer := range q.consumerBuffers {
		qv := &bufferValue{
			value:      value,
			producerId: id,
		}
		buffer.rwm.RLock()
		root := buffer.root
		buffer.rwm.RUnlock()
		if root == nil {
			buffer.rwm.Lock()
			buffer.root = qv
			buffer.rwm.Unlock()
		} else {
			buffer.rwm.RLock()
			for root.next != nil {
				root = root.next
			}
			buffer.rwm.RUnlock()
			buffer.rwm.Lock()
			root.next = qv
			buffer.rwm.Unlock()
		}
		buffer.c.Signal()
	}
	q.rLocker.Unlock()
	return nil
}

func (q *queue) consume(id string) (interface{}, error) {
	q.rLocker.Lock()
	buffer, isPresent := q.consumerBuffers[id]
	q.rLocker.Unlock()
	if !isPresent {
		return nil, fmt.Errorf("unknown consumer")
	}
	buffer.rwm.Lock()
	if buffer.root == nil {
		buffer.c.Wait()
	}
	root := buffer.root
	buffer.root = buffer.root.next
	buffer.rwm.Unlock()
	return root.value, nil
}

func (q *queue) cancel(id string) {
	q.wLocker.Lock()
	delete(q.consumerBuffers, id)
	q.wLocker.Unlock()
}

func (q *queue) NewProducer() (Producer, error) {
	return &producer{
		id: ksuid.New().String(),
		q:  q,
	}, nil
}

func (q *queue) NewConsumer() (Consumer, error) {
	id := ksuid.New().String()
	rwm := &sync.RWMutex{}
	q.wLocker.Lock()
	q.consumerBuffers[id] = &linkedBuffer{
		root: nil,
		rwm:  rwm,
		c:    sync.NewCond(rwm),
	}
	q.wLocker.Unlock()
	return &consumer{
		id: id,
		q:  q,
	}, nil
}

func NewQueue() Queue {
	q := newQueue()
	return q
}
