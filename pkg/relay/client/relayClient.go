package client

import (
	"github.com/raf924/bot/pkg/config/bot"
	"github.com/raf924/bot/pkg/queue"
	messages "github.com/raf924/connector-api/pkg/gen"
)

var connectorRelays = map[string]RelayBuilder{}

type RelayBuilder func(config interface{}, withBotExchange *queue.Exchange) RelayClient

func RegisterRelayClient(key string, relayBuilder RelayBuilder) {
	connectorRelays[key] = relayBuilder
}

func GetRelayClient(config bot.Config, botExchange *queue.Exchange) RelayClient {
	for key, config := range config.Connector {
		if builder, ok := connectorRelays[key]; ok {
			return builder(config, botExchange)
		}
	}
	return nil
}

type RelayClient interface {
	GetUsers() []*messages.User
	OnUserJoin(func(user *messages.User, timestamp int64))
	OnUserLeft(func(user *messages.User, timestamp int64))
	Connect(registration *messages.RegistrationPacket) (*messages.User, error)
	Done() <-chan struct{}
}
