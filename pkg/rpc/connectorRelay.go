package rpc

import (
	"context"
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/domain"
)

var connectorRelayBuilders = map[string]ConnectorRelayBuilder{}

type ConnectorRelayBuilder func(config interface{}) ConnectorRelay

func RegisterConnectorRelay(name string, relayBuilder ConnectorRelayBuilder) {
	connectorRelayBuilders[name] = relayBuilder
}

var _ = RegisterConnectorRelay

func GetConnectorRelay(config connector.Config) ConnectorRelay {
	for key, config := range config.Bot {
		if builder, ok := connectorRelayBuilders[key]; ok {
			return builder(config)
		}
	}
	return nil
}

type ConnectorRelay interface {
	Start(ctx context.Context, botUser *domain.User, onlineUsers domain.UserList, trigger string) error
	Accept() (Dispatcher, error)
	Recv() (*domain.ClientMessage, error)
	Done() <-chan struct{}
	Err() error
}
