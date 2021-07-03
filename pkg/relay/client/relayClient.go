package client

import (
	"github.com/raf924/bot/pkg/config/bot"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/proto"
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
	GetUsers() []*messages.User
	OnUserJoin(func(user *messages.User, timestamp int64))
	OnUserLeft(func(user *messages.User, timestamp int64))
	Connect(registration *messages.RegistrationPacket) (*messages.User, error)
	Send(packet *messages.BotPacket) error
	Recv() (proto.Message, error)
	Done() <-chan struct{}
}
