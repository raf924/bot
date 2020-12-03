package relay

import (
	"github.com/raf924/bot/api/messages"
	"github.com/raf924/bot/pkg/config/bot"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var connectorRelays = map[string]RelayBuilder{}

type RelayBuilder func(config interface{}) ConnectorRelay

func RegisterConnectorRelay(key string, relayBuilder RelayBuilder) {
	connectorRelays[key] = relayBuilder
}

func GetConnectorRelay(config bot.Config) ConnectorRelay {
	for key, config := range config.Connector {
		if builder, ok := connectorRelays[key]; ok {
			return builder(config)
		}
	}
	return nil
}

type ConnectorMessage struct {
	Message protoreflect.ProtoMessage
}

type ConnectorRelay interface {
	GetUsers() []*messages.User
	OnUserJoin(func(user *messages.User, timestamp int64))
	OnUserLeft(func(user *messages.User, timestamp int64))
	Connect(registration *messages.RegistrationPacket) (*messages.User, error)
	Send(message *messages.BotPacket) error
	Recv() (*ConnectorMessage, error)
	RecvMsg(packet *ConnectorMessage) error
	Done() <-chan struct{}
}
