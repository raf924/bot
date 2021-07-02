package queue

import (
	"github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type valueGenerator interface {
	Gen() interface{}
}

type valueGeneratorFunc func() interface{}

func (v valueGeneratorFunc) Gen() interface{} {
	return v()
}

type typeBenchmark struct {
	vGen valueGenerator
}

var intBenchmark = typeBenchmark{
	vGen: valueGeneratorFunc(func() interface{} {
		return 5
	}),
}

var messagePacketBenchmark = typeBenchmark{
	vGen: valueGeneratorFunc(func() interface{} {
		return &gen.MessagePacket{
			Timestamp: timestamppb.New(time.Now()),
			Message:   "Hello there",
			User: &gen.User{
				Nick:  "Test",
				Id:    "test",
				Mod:   false,
				Admin: false,
			},
			Private: false,
		}
	}),
}

func TestProtoMessage(t *testing.T) {
	var q = NewQueue()
	p, _ := q.NewProducer()
	c, _ := q.NewConsumer()
	_ = p.Produce(&gen.MessagePacket{
		Timestamp: timestamppb.Now(),
		Message:   "test",
		User: &gen.User{
			Nick:  "user",
			Id:    "id",
			Mod:   false,
			Admin: false,
		},
		Private: false,
	})
	var pm interface{}
	pm, _ = c.Consume()
	switch pm.(type) {
	case *gen.MessagePacket:
	default:
		t.Errorf("expected MessagePacket, got %v", pm)
	}
}

func TestQueueConsumer_Consume(t *testing.T) {
	q := NewQueue()
	p, err := q.NewProducer()
	if err != nil {
		t.Errorf("unexpected error = %v", err)
	}
	c, err := q.NewConsumer()
	if err != nil {
		t.Errorf("unexpected error = %v", err)
		return
	}
	err = p.Produce(5)
	if err != nil {
		t.Errorf("unexpected error = %v", err)
		return
	}
	i, err := c.Consume()
	if err != nil {
		t.Errorf("unexpected error = %v", err)
		return
	}
	if i.(int) != 5 {
		t.Errorf("expected %v got %v", 5, i)
		return
	}
}

func benchmarkType(valueGen valueGenerator, valueCount int) func(b *testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		q := NewQueue()
		p, err := q.NewProducer()
		if err != nil {
			b.Errorf("unexpected error = %v", err)
			return
		}
		c, err := q.NewConsumer()
		if err != nil {
			b.Errorf("unexpected error = %v", err)
			return
		}
		for i := 0; i < b.N; i++ {
			for i := 0; i < valueCount; i++ {
				err := p.Produce(valueGen.Gen())
				if err != nil {
					b.Errorf("unexpected error = %v", err)
				}
			}
			wg := sync.WaitGroup{}
			wg.Add(valueCount)
			b.StartTimer()
			var r interface{}
			for i := 0; i < valueCount; i++ {
				go func() {
					if r, err = c.Consume(); err != nil {
						b.Errorf("unexpected error = %v", err)
					}
					wg.Done()
				}()
			}
			wg.Wait()
			b.StopTimer()
			_ = r
		}
	}
}

func TestConsumer_Cancel(t *testing.T) {
	q := NewQueue()
	p, _ := q.NewProducer()
	c, _ := q.NewConsumer()
	_ = p.Produce(5)
	c.Cancel()
	_, err := c.Consume()
	if err == nil {
		t.Errorf("expected error")
	}
}

func BenchmarkQueue(b *testing.B) {
	for i := 1; i < b.N; i++ {
		b.Run("Consume Ints", benchmarkType(intBenchmark.vGen, i))
	}
}

func BenchmarkProducer_Produce(b *testing.B) {
	b.ReportAllocs()
	q := NewQueue()
	c, _ := q.NewConsumer()
	for i := 0; i < b.N; i++ {
		err := q.produce("", 5)
		if err != nil {
			b.Error(err)
		}
		_, err = c.Consume()
		if err != nil {
			b.Error(err)
		}
	}
}

func TestConsumer_Consume(t *testing.T) {
	var q = NewQueue()
	c1, _ := q.NewConsumer()
	c2, _ := q.NewConsumer()
	c := make(chan interface{})
	go func() {
		v, _ := c1.Consume()
		c <- v
	}()
	go func() {
		v, _ := c2.Consume()
		c <- v
	}()
	p, _ := q.NewProducer()
	err := p.Produce(4)
	if err != nil {
		t.Error(err)
	}
	v1 := <-c
	v2 := <-c
	if v1 != 4 && v1 != v2 {
		t.Errorf("expected %v got %v and %v", 4, v1, v2)
	}
}

func TestQueueFifo(t *testing.T) {
	var valueCount = 10
	var values = make([]int, valueCount)
	for i := 0; i < len(values); i++ {
		values[i] = rand.Int()
	}
	var q = NewQueue()
	c, _ := q.NewConsumer()
	p, _ := q.NewProducer()
	for _, value := range values {
		if err := p.Produce(value); err != nil {
			t.Fatalf("unexpected error = %v", err)
		}
	}
	for i := 0; i < valueCount; i++ {
		v, err := c.Consume()
		if err != nil {
			t.Fatalf("unexpected err = %v", err)
		}
		if v != values[i] {
			t.Fatalf("expected %v got %v", values[i], v)
		}
	}
}

func TestQueueSequential(t *testing.T) {
	q := NewQueue()
	p, _ := q.NewProducer()
	c, _ := q.NewConsumer()
	_ = p.Produce(5)
	_, _ = c.Consume()
	_ = p.Produce(6)
	_, _ = c.Consume()
}

func TestParallelConsumer(t *testing.T) {
	q := NewQueue()
	p, _ := q.NewProducer()
	c, _ := q.NewConsumer()
	go func() {
		_, _ = c.Consume()
		_, _ = c.Consume()
	}()
	time.Sleep(500 * time.Microsecond)
	_ = p.Produce(5)
	_ = p.Produce(6)
}
