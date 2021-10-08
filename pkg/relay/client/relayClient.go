package client

import (
	"github.com/raf924/bot/pkg/config/bot"
	"github.com/raf924/bot/pkg/domain"
)

var connectorRelays = map[string]RelayBuilder{}

type RelayBuilder func(config interface{}) RelayClient

func RegisterRelayClient(key string, relayBuilder RelayBuilder) {
	connectorRelays[key] = relayBuilder
}

func GetRelayClient(config bot.Config) RelayClient {
	for key, config := range config.Connector {
		if builder, ok := connectorRelays[key]; ok {
			return builder(config)
		}
	}
	return nil
}

type RelayClient interface {
	Connect(registration *domain.RegistrationMessage) (*domain.User, error)
	Send(packet *domain.ClientMessage) error
	Recv() (domain.ServerMessage, error)
}
