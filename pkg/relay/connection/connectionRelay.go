package connection

import (
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/domain"
	"time"
)

var connectionRelays = map[string]RelayBuilder{}

type RelayBuilder func(config interface{}) Relay

func RegisterConnectionRelay(key string, relayBuilder RelayBuilder) {
	connectionRelays[key] = relayBuilder
}

func GetConnectionRelay(config connector.Config) Relay {
	for key, config := range config.Connection {
		if builder, ok := connectionRelays[key]; ok {
			return builder(config)
		}
	}
	return nil
}

type Message interface {
	unimplementedMessageMethod()
}

type emptyMessage struct{}

func (e emptyMessage) unimplementedMessageMethod() {}

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

type Relay interface {
	Recv() (*domain.ChatMessage, error)
	Send(message *domain.ClientMessage) error
	OnUserJoin(func(user *domain.User, timestamp time.Time))
	OnUserLeft(func(user *domain.User, timestamp time.Time))
	Connect(nick string) (*domain.User, domain.UserList, error)
}
