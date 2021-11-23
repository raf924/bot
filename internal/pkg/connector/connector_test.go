package connector

import (
	"context"
	"fmt"
	internalRpc "github.com/raf924/bot/v2/internal/pkg/rpc"
	"github.com/raf924/bot/v2/pkg"
	"github.com/raf924/bot/v2/pkg/config/connector"
	"github.com/raf924/connector-sdk/domain"
	"github.com/raf924/connector-sdk/queue"
	"testing"
	"time"
)

type dummyConnection struct {
	users                 domain.UserList
	chatMessageConsumer   queue.Consumer
	clientMessageProducer queue.Producer
	botUser               *domain.User
}

func (d *dummyConnection) Recv() (*domain.ChatMessage, error) {
	consume, err := d.chatMessageConsumer.Consume()
	if err != nil {
		return nil, err
	}
	if _, ok := consume.(*domain.ChatMessage); !ok {
		return nil, fmt.Errorf("")
	}
	return consume.(*domain.ChatMessage), nil
}

func (d *dummyConnection) Send(message *domain.ClientMessage) error {
	return d.clientMessageProducer.Produce(message)
}

func (d *dummyConnection) OnUserJoin(func(user *domain.User, timestamp time.Time)) {

}

func (d *dummyConnection) OnUserLeft(func(user *domain.User, timestamp time.Time)) {

}

func (d *dummyConnection) Connect(nick string) (*domain.User, domain.UserList, error) {
	return domain.NewUser(nick, "", domain.RegularUser), d.users, nil
}

type dummyRunnable struct {
	ctx context.Context
}

func (d *dummyRunnable) Start(ctx context.Context) error {
	d.ctx = ctx
	return nil
}

func (d *dummyRunnable) Done() <-chan struct{} {
	return d.ctx.Done()
}

func (d *dummyRunnable) Err() error {
	return d.ctx.Err()
}

var _ pkg.Runnable = (*dummyRunnable)(nil)

func TestConnector(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	clientMessageQueue := queue.NewQueue()
	chatMessageQueue := queue.NewQueue()
	serverMessageQueue := queue.NewQueue()
	clientMessageConsumer, err := clientMessageQueue.NewConsumer()
	if err != nil {
		t.Fatal(err)
	}
	clientMessageProducer, err := clientMessageQueue.NewProducer()
	if err != nil {
		t.Fatal(err)
	}
	chatMessageConsumer, err := chatMessageQueue.NewConsumer()
	if err != nil {
		t.Fatal(err)
	}
	chatMessageProducer, err := chatMessageQueue.NewProducer()
	if err != nil {
		t.Fatal(err)
	}
	serverMessageConsumer, err := serverMessageQueue.NewConsumer()
	if err != nil {
		t.Fatal(err)
	}
	serverMessageProducer, err := serverMessageQueue.NewProducer()
	if err != nil {
		t.Fatal(err)
	}
	botUser := domain.NewOnlineUser("bot", "id", domain.RegularUser, time.Now())
	crRelay := internalRpc.NewDefaultConnectorRelay(
		&dummyRunnable{ctx: ctx},
		clientMessageConsumer,
		serverMessageProducer,
	)
	cnRelay := &dummyConnection{
		users:                 domain.NewUserList(botUser),
		botUser:               botUser,
		chatMessageConsumer:   chatMessageConsumer,
		clientMessageProducer: clientMessageProducer,
	}
	ctr := NewConnector(connector.Config{}, cnRelay, crRelay)
	err = ctr.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	message := domain.NewChatMessage("hello", botUser, nil, false, false, time.Now(), true)
	err = chatMessageProducer.Produce(message)
	if err != nil {
		t.Fatal(err)
	}
	consume, err := serverMessageConsumer.Consume()
	if err != nil {
		t.Fatal(err)
	}
	if consume != message {
		t.Fatal("expected", message, "got", consume)
	}
}
