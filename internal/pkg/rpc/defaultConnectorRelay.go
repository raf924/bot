package rpc

import (
	"context"
	"github.com/raf924/bot/v2/pkg"
	"github.com/raf924/connector-sdk/domain"
	"github.com/raf924/connector-sdk/queue"
	"github.com/raf924/connector-sdk/rpc"
)

type defaultConnectorRelay struct {
	ctx                   context.Context
	accepted              bool
	bot                   pkg.Runnable
	clientMessageConsumer queue.Consumer
	serverMessageProducer queue.Producer
}

func (d *defaultConnectorRelay) Start(ctx context.Context, _ *domain.User, _ domain.UserList, _ string) error {
	d.ctx = ctx
	err := d.bot.Start(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *defaultConnectorRelay) Accept() (rpc.Dispatcher, error) {
	if d.accepted {
		<-d.ctx.Done()
		return nil, d.ctx.Err()
	}
	d.accepted = true
	return &defaultDispatcher{
		ctx:                   d.ctx,
		commands:              domain.NewCommandList(),
		serverMessageProducer: d.serverMessageProducer,
	}, nil
}

func (d *defaultConnectorRelay) Recv() (*domain.ClientMessage, error) {
	consume, err := d.clientMessageConsumer.Consume()
	if err != nil {
		return nil, err
	}
	return consume.(*domain.ClientMessage), nil
}

func (d *defaultConnectorRelay) Done() <-chan struct{} {
	return d.ctx.Done()
}

func (d *defaultConnectorRelay) Err() error {
	return d.ctx.Err()
}

var _ rpc.ConnectorRelay = (*defaultConnectorRelay)(nil)

func NewDefaultConnectorRelay(runnable pkg.Runnable, clientMessageConsumer queue.Consumer, serverMessageProducer queue.Producer) rpc.ConnectorRelay {
	return &defaultConnectorRelay{
		bot:                   runnable,
		accepted:              false,
		serverMessageProducer: serverMessageProducer,
		clientMessageConsumer: clientMessageConsumer,
	}
}
