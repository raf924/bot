package rpc

import (
	"context"
	"github.com/raf924/connector-sdk/domain"
	"github.com/raf924/connector-sdk/rpc"
	"github.com/raf924/queue"
)

type defaultDispatcherRelay struct {
	ctx                   context.Context
	onlineUsers           domain.UserList
	trigger               string
	currentUser           *domain.User
	clientMessageProducer queue.Producer[*domain.ClientMessage]
	serverMessageConsumer queue.Consumer[domain.ServerMessage]
}

func (d *defaultDispatcherRelay) Connect(*domain.RegistrationMessage) (*domain.ConfirmationMessage, error) {
	return domain.NewConfirmationMessage(d.currentUser, d.trigger, d.onlineUsers.All()), nil
}

func (d *defaultDispatcherRelay) Send(packet *domain.ClientMessage) error {
	return d.clientMessageProducer.Produce(packet)
}

func (d *defaultDispatcherRelay) Recv() (domain.ServerMessage, error) {
	return d.serverMessageConsumer.Consume(d.ctx)
}

func (d *defaultDispatcherRelay) Done() <-chan struct{} {
	return d.ctx.Done()
}

func (d *defaultDispatcherRelay) Err() error {
	return d.ctx.Err()
}

var _ rpc.DispatcherRelay = (*defaultDispatcherRelay)(nil)

func NewDefaultDispatcherRelay(ctx context.Context, onlineUsers domain.UserList, trigger string, currentUser *domain.User, clientMessageProducer queue.Producer[*domain.ClientMessage], serverMessageConsumer queue.Consumer[domain.ServerMessage]) rpc.DispatcherRelay {
	return &defaultDispatcherRelay{
		ctx:                   ctx,
		onlineUsers:           onlineUsers,
		trigger:               trigger,
		currentUser:           currentUser,
		clientMessageProducer: clientMessageProducer,
		serverMessageConsumer: serverMessageConsumer,
	}
}
