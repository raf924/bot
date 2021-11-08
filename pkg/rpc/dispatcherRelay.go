package rpc

import (
	"github.com/raf924/bot/pkg/config/bot"
	"github.com/raf924/bot/pkg/domain"
)

var dispatcherRelayBuilders = map[string]DispatcherRelayBuilder{}

type DispatcherRelayBuilder func(config interface{}) DispatcherRelay

func RegisterDispatcherRelay(key string, relayBuilder DispatcherRelayBuilder) {
	dispatcherRelayBuilders[key] = relayBuilder
}

var _ = RegisterDispatcherRelay

func GetDispatcherRelay(config bot.Config) DispatcherRelay {
	for key, config := range config.Connector {
		if builder, ok := dispatcherRelayBuilders[key]; ok {
			return builder(config)
		}
	}
	return nil
}

type DispatcherRelay interface {
	Connect(registration *domain.RegistrationMessage) (*domain.ConfirmationMessage, error)
	Send(packet *domain.ClientMessage) error
	Recv() (domain.ServerMessage, error)
	Done() <-chan struct{}
	Err() error
}
