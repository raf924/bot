package server

import (
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/queue"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/proto"
)

var relayServers = map[string]RelayServerBuilder{}

type RelayServerBuilder func(config interface{}, connectorExchange *queue.Exchange) RelayServer

func RegisterRelayServer(name string, relayBuilder RelayServerBuilder) {
	relayServers[name] = relayBuilder
}

func GetRelayServer(config connector.Config, connectorExchange *queue.Exchange) RelayServer {
	for key, config := range config.Bot {
		if builder, ok := relayServers[key]; ok {
			return builder(config, connectorExchange)
		}
	}
	return nil
}

type RelayServer interface {
	Start(botUser *messages.User, users []*messages.User, trigger string) error
	Commands() []*messages.Command
	Send(message proto.Message) error
	Recv() (*messages.BotPacket, error)
}
