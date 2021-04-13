package commands

import (
	"github.com/raf924/bot/pkg/bot/command"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

type EchoCommand struct {
	//command.NoOpCommand
	command.NoOpInterceptor
}

func (e *EchoCommand) Init(bot command.Executor) error {
	return nil
}

func (e *EchoCommand) Name() string {
	return "echo"
}

func (e *EchoCommand) Aliases() []string {
	return []string{"e"}
}

func (e *EchoCommand) Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	packet := &messages.BotPacket{
		Timestamp: timestamppb.New(time.Now()),
		Message:   strings.Join(command.Args, " "),
		Recipient: nil,
		Private:   false,
	}
	return []*messages.BotPacket{packet}, nil
}
