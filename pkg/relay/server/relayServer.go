package server

import (
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/domain"
)

var relayServers = map[string]RelayServerBuilder{}

type RelayServerBuilder func(config interface{}) RelayServer

func RegisterRelayServer(name string, relayBuilder RelayServerBuilder) {
	relayServers[name] = relayBuilder
}

func GetRelayServer(config connector.Config) RelayServer {
	for key, config := range config.Bot {
		if builder, ok := relayServers[key]; ok {
			return builder(config)
		}
	}
	return nil
}

type RelayServer interface {
	Start(botUser *domain.User, onlineUsers domain.UserList, trigger string) error
	Commands() domain.CommandList
	Send(message domain.ServerMessage) error
	Recv() (*domain.ClientMessage, error)
}
