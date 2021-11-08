package rpc

import (
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/domain"
	"time"
)

var connectionRelayBuilders = map[string]ConnectionRelayBuilder{}

type ConnectionRelayBuilder func(config interface{}) ConnectionRelay

func RegisterConnectionRelay(key string, relayBuilder ConnectionRelayBuilder) {
	connectionRelayBuilders[key] = relayBuilder
}

var _ = RegisterConnectionRelay

func GetConnectionRelay(config connector.Config) ConnectionRelay {
	for key, config := range config.Connection {
		if builder, ok := connectionRelayBuilders[key]; ok {
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

type ConnectionRelay interface {
	Recv() (*domain.ChatMessage, error)
	Send(message *domain.ClientMessage) error
	OnUserJoin(func(user *domain.User, timestamp time.Time))
	OnUserLeft(func(user *domain.User, timestamp time.Time))
	Connect(nick string) (*domain.User, domain.UserList, error)
}
