package bot

import (
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/bot/permissions"
	"github.com/raf924/bot/pkg/domain"
)

type CommandHandler struct {
	commands                 domain.CommandList
	loadedCommands           map[string]command.Command
	botUser                  *domain.User
	commandCallback          func([]*domain.ClientMessage, error) error
	userPermissionManager    permissions.PermissionManager
	commandPermissionManager permissions.PermissionManager
}

func (c *CommandHandler) PassServerMessage(message domain.ServerMessage, senderIsBanned bool) error {
	if message, ok := message.(*domain.UserEvent); ok {
		for _, cmd := range c.commands.All() {
			var chatInterceptor command.Interceptor = c.loadedCommands[cmd.Name()]
			err := c.commandCallback(chatInterceptor.OnUserEvent(message))
			if err != nil {
				return err
			}
		}
		return nil
	}
	var sender = message.(FromUser).Sender()
	switch message := message.(type) {
	case *domain.ChatMessage:
		for _, cmd := range c.commands.All() {
			if !c.isAllowed(cmd.Name(), sender) {
				continue
			}
			var chatInterceptor command.Interceptor = c.loadedCommands[cmd.Name()]
			if !message.Incoming() && chatInterceptor.IgnoreSelf() {
				continue
			}
			err := c.commandCallback(chatInterceptor.OnChat(message))
			if err != nil {
				return err
			}
		}
	case *domain.CommandMessage:
		if senderIsBanned {
			return nil
		}
		cmd := c.commands.Find(message.Command())
		if cmd == nil {
			return c.PassServerMessage(message.ToChatMessage(), senderIsBanned)
		}
		if !c.isAllowed(cmd.Name(), sender) {
			return nil
		}
		var executable = c.loadedCommands[cmd.Name()]
		if !message.Sender().Is(c.botUser) && executable.IgnoreSelf() {
			return nil
		}
		err := c.commandCallback(executable.Execute(message))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CommandHandler) isAllowed(command string, user *domain.User) bool {
	uPermission, err := c.userPermissionManager.GetPermission(user.Id())
	if err != nil {
		return false
	}
	cPermission, err := c.commandPermissionManager.GetPermission(command)
	if err != nil {
		return false
	}
	return uPermission.Has(cPermission)
}
