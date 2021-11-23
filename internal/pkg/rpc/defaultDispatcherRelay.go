package rpc

import (
	"context"
	"github.com/raf924/connector-sdk/domain"
	"github.com/raf924/connector-sdk/queue"
	"github.com/raf924/connector-sdk/rpc"
)

type defaultDispatcherRelay struct {
	ctx                   context.Context
	onlineUsers           domain.UserList
	trigger               string
	currentUser           *domain.User
	clientMessageProducer queue.Producer
	serverMessageConsumer queue.Consumer
}

func (d *defaultDispatcherRelay) Connect(*domain.RegistrationMessage) (*domain.ConfirmationMessage, error) {
	return domain.NewConfirmationMessage(d.currentUser, d.trigger, d.onlineUsers.All()), nil
}

func (d *defaultDispatcherRelay) Send(packet *domain.ClientMessage) error {
	return d.clientMessageProducer.Produce(packet)
}

func (d *defaultDispatcherRelay) Recv() (domain.ServerMessage, error) {
	consume, err := d.serverMessageConsumer.Consume()
	if err != nil {
		return nil, err
	}
	return consume.(domain.ServerMessage), nil
}

func (d *defaultDispatcherRelay) Done() <-chan struct{} {
	return d.ctx.Done()
}

func (d *defaultDispatcherRelay) Err() error {
	return d.ctx.Err()
}

var _ rpc.DispatcherRelay = (*defaultDispatcherRelay)(nil)

func NewDefaultDispatcherRelay(ctx context.Context, onlineUsers domain.UserList, trigger string, currentUser *domain.User, clientMessageProducer queue.Producer, serverMessageConsumer queue.Consumer) rpc.DispatcherRelay {
	return &defaultDispatcherRelay{
		ctx:                   ctx,
		onlineUsers:           onlineUsers,
		trigger:               trigger,
		currentUser:           currentUser,
		clientMessageProducer: clientMessageProducer,
		serverMessageConsumer: serverMessageConsumer,
	}
}
