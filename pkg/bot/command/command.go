package command

import (
	"github.com/raf924/bot/pkg/bot/permissions"
	messages "github.com/raf924/connector-api/pkg/gen"
)

type Executor interface {
	BotUser() *messages.User
	ApiKeys() map[string]string
	OnlineUsers() map[string]messages.User
	UserHasPermission(user *messages.User, permission permissions.Permission) bool
}

type Interceptor interface {
	// OnChat should be implemented if the command needs to handle chat messages as they arrive
	OnChat(message *messages.MessagePacket) ([]*messages.BotPacket, error)
	// OnUserEvent should be implemented if the command needs to handle user events
	OnUserEvent(packet *messages.UserPacket) ([]*messages.BotPacket, error)
	// if IgnoreSelf no message sent by the bot is send to the command
	IgnoreSelf() bool
}

type Executable interface {
	// Init should be implemented to access external data such as API keys and user list or even the bot's nick
	Init(bot Executor) error
	// Name must be implemented for command recognition. It must return a unique alphanumerical string compliant with the following regex: /^[a-z]([0-9]|[a-z])*$/
	//
	//You have to know what other commands exist to avoid overlap
	Name() string
	// Aliases should return a list of command aliases (excluding the string returned by Name) to be used in command recognition
	Aliases() []string
	// Execute is the core function of a command. It does the work when called.
	Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error)
}

// A Command can either be triggered by its Name or Aliases with arguments or by chat events.
// The triggered methods are not required to return anything.
type Command interface {
	Executable
	Interceptor
}

// NoOpCommand : Commands should embed NoOpCommand to avoid noop method implementations.
// By embedding this, a Command implementation only needs to implement Name and Execute for basic functionality
// IgnoreSelf returns true
type NoOpCommand struct {
}

func (n *NoOpCommand) Init(_ Executor) error {
	return nil
}

func (n *NoOpCommand) Name() string {
	panic("implement me")
}

func (n *NoOpCommand) Aliases() []string {
	return []string{}
}

func (n *NoOpCommand) Execute(_ *messages.CommandPacket) ([]*messages.BotPacket, error) {
	panic("implement me")
}

func (n *NoOpCommand) OnChat(_ *messages.MessagePacket) ([]*messages.BotPacket, error) {
	return nil, nil
}

func (n *NoOpCommand) OnUserEvent(_ *messages.UserPacket) ([]*messages.BotPacket, error) {
	return nil, nil
}

func (n *NoOpCommand) IgnoreSelf() bool {
	return true
}

// Commands should embed NoOpInterceptor to avoid noop method implementations.
// By embedding this, a Command implementation only needs to implement the Executable interface
// IgnoreSelf returns true by default
type NoOpInterceptor struct {
}

func (n *NoOpInterceptor) OnChat(_ *messages.MessagePacket) ([]*messages.BotPacket, error) {
	return nil, nil
}

func (n *NoOpInterceptor) OnUserEvent(_ *messages.UserPacket) ([]*messages.BotPacket, error) {
	return nil, nil
}

func (n *NoOpInterceptor) IgnoreSelf() bool {
	return true
}
