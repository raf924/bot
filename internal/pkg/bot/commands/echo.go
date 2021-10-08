package commands

import (
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/domain"
	"strings"
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

func (e *EchoCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	packet := domain.NewClientMessage(strings.Join(command.Args(), " "), nil, command.Private())
	return []*domain.ClientMessage{packet}, nil
}
