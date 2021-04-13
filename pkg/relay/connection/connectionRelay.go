package connection

import (
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/queue"
	messages "github.com/raf924/connector-api/pkg/gen"
)

var connectionRelays = map[string]ConnectionRelayBuilder{}

type ConnectionRelayBuilder func(config interface{}, connectorExchange *queue.Exchange) ConnectionRelay

func RegisterConnectionRelay(key string, relayBuilder ConnectionRelayBuilder) {
	connectionRelays[key] = relayBuilder
}

func GetConnectionRelay(config connector.Config, connectorExchange *queue.Exchange) ConnectionRelay {
	for key, config := range config.Connection {
		if builder, ok := connectionRelays[key]; ok {
			return builder(config, connectorExchange)
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
}
