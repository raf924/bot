package rpc

import (
	"context"
	"github.com/raf924/bot/pkg/domain"
	"github.com/raf924/bot/pkg/queue"
	"github.com/raf924/bot/pkg/rpc"
)

type defaultDispatcher struct {
	ctx                   context.Context
	commands              domain.CommandList
	serverMessageProducer queue.Producer
}

func (d *defaultDispatcher) Dispatch(message domain.ServerMessage) error {
	return d.serverMessageProducer.Produce(message)
}

func (d *defaultDispatcher) Commands() domain.CommandList {
	return d.commands
}

func (d *defaultDispatcher) Done() <-chan struct{} {
	return d.ctx.Done()
}

func (d *defaultDispatcher) Err() error {
	return d.ctx.Err()
}

var _ rpc.Dispatcher = (*defaultDispatcher)(nil)