package relay

import (
	"github.com/raf924/bot/pkg/config/connector"
	messages "github.com/raf924/connector-api/pkg/gen"
)

var botRelays = map[string]BotRelayBuilder{}

type BotRelayBuilder func(config interface{}) BotRelay

func RegisterBotRelay(name string, relayBuilder BotRelayBuilder) {
	botRelays[name] = relayBuilder
}

func GetBotRelay(config connector.Config) BotRelay {
	for key, config := range config.Connector.Bot {
		if builder, ok := botRelays[key]; ok {
			return builder(config)
		}
	}
	return nil
}

type BotRelay interface {
	Start(botUser *messages.User, users []*messages.User) error
	PassMessage(message *messages.MessagePacket) error
	PassEvent(event *messages.UserPacket) error
	PassCommand(command *messages.CommandPacket) error
	RecvMsg(packet *messages.BotPacket) error
	Trigger() string
	Commands() []*messages.Command
	Ready() <-chan struct{}
}
