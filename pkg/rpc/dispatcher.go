package rpc

import "github.com/raf924/bot/pkg/domain"

type Dispatcher interface {
	Dispatch(message domain.ServerMessage) error
	Commands() domain.CommandList
	Done() <-chan struct{}
	Err() error
}
