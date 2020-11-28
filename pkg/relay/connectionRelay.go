package relay

import (
	"github.com/raf924/bot/api/messages"
	"github.com/raf924/bot/pkg/config/connector"
)

var connectionRelays = map[string]ConnectionRelayBuilder{}

type ConnectionRelayBuilder func(config interface{}) ConnectionRelay

func RegisterConnectionRelay(key string, relayBuilder ConnectionRelayBuilder) {
	connectionRelays[key] = relayBuilder
}

func GetConnectionRelay(config connector.Config) ConnectionRelay {
	for key, config := range config.Connector.Connection {
		if builder, ok := connectionRelays[key]; ok {
			return builder(config)
		}
	}
	return nil
}

type Message interface {
	unimplementedMethod()
}

type emptyMessage struct{}

func (e emptyMessage) unimplementedMethod() {}

type ChatMessage struct {
	emptyMessage
	Message   string
	Recipient string
	Private   bool
}

type NoticeMessage struct {
	emptyMessage
	Message string
}

type InviteMessage struct {
	emptyMessage
	Recipient string
}

type ConnectionRelay interface {
	CommandTrigger() string
	GetUsers() []*messages.User
	OnUserJoin(func(user *messages.User, timestamp int64))
	OnUserLeft(func(user *messages.User, timestamp int64))
	Connect(nick string) error
	Send(message Message) error
	Recv() (*messages.MessagePacket, error)
	RecvMsg(packet *messages.MessagePacket) error
}
