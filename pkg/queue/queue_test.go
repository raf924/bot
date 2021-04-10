package queue

import (
	"fmt"
	"github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
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

type typeConsumer func(consumer *Consumer) error

type typeBenchmark struct {
	typ      reflect.Type
	vw       ValueWriter
	vGen     valueGenerator
	consumer typeConsumer
}

var intBenchmark = typeBenchmark{
	typ: reflect.TypeOf(5),
	vw: ValueWriterFunc(func(ptr, value interface{}) (err error) {
		defer func() {
			r := recover()
			if r != nil {
				err = fmt.Errorf("error: %v", r)
			}
		}()
		*(ptr.(*int)) = value.(int)
		return nil
	}),
	vGen: valueGeneratorFunc(func() interface{} {
		return 5
	}),
	consumer: func(consumer *Consumer) error {
		var i int
		return consumer.Consume(&i)
	},
}

var messagePacketBenchmark = typeBenchmark{
	typ: reflect.TypeOf(&gen.MessagePacket{}),
	vw: ValueWriterFunc(func(ptr, value interface{}) (err error) {
		defer func() {
			r := recover()
			if r != nil {
				err = fmt.Errorf("error: %v", r)
			}
		}()
		pm := ptr.(*gen.MessagePacket)
		m := value.(*gen.MessagePacket)
		pm.Private = m.GetPrivate()
		pm.User = m.GetUser()
		pm.Message = m.GetMessage()
		pm.Timestamp = m.GetTimestamp()
		return nil
	}),
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
	consumer: func(consumer *Consumer) error {
		var m gen.MessagePacket
		return consumer.Consume(&m)
	},
}

func TestQueueConsumer_Consume(t *testing.T) {
	q := NewQueue(WithValueWriter(intBenchmark.vw))
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
	var i int
	err = c.Consume(&i)
	if err != nil {
		t.Errorf("unexpected error = %v", err)
		return
	}
	if i != 5 {
		t.Errorf("expected %v got %v", 5, i)
		return
	}
}

func benchmarkValueWriter(vw ValueWriter, valueGen valueGenerator, consume typeConsumer) func(b *testing.B) {
	return func(b *testing.B) {
		q := NewQueue(WithValueWriter(vw))
		p, err := q.NewProducer()
		if err != nil {
			b.Errorf("unexpected error = %v", err)
		}
		c, err := q.NewConsumer()
		if err != nil {
			b.Errorf("unexpected error = %v", err)
			return
		}
		for i := 0; i < b.N; i++ {
			err := p.Produce(valueGen.Gen())
			if err != nil {
				b.Errorf("unexpected error = %v", err)
				return
			}
		}
		b.Run("Consume", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := consume(c); err != nil {
					b.Errorf("unexpected error = %v", err)
					return
				}
			}
		})
	}
}

func benchmarkConsumeType(benchmark typeBenchmark) func(b *testing.B) {
	return func(b *testing.B) {
		b.Run("Type value writer", benchmarkValueWriter(benchmark.vw, benchmark.vGen, benchmark.consumer))
		b.Run("Reflect value writer", benchmarkValueWriter(defaultVW, benchmark.vGen, benchmark.consumer))
	}
}

func BenchmarkQueueConsumer_Consume(b *testing.B) {
	b.Run("Consume Ints", benchmarkConsumeType(intBenchmark))
	b.Run("Consume MessagePackets", benchmarkConsumeType(messagePacketBenchmark))
}
