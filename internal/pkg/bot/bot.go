package bot

import (
	"context"
	"fmt"
	"github.com/raf924/bot/api/messages"
	_ "github.com/raf924/bot/internal/pkg/bot/permissions"
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/bot/permissions"
	"github.com/raf924/bot/pkg/config/bot"
	"github.com/raf924/bot/pkg/relay"
	"log"
	"time"
)

const defaultTimeout = 30 * time.Second

var Commands []command.Command

type ban struct {
	start    time.Time
	duration time.Duration
}

type Bot struct {
	connectorRelay           relay.ConnectorRelay
	commands                 map[string]command.Command
	config                   bot.Config
	botUser                  *messages.User
	userPermissionManager    permissions.PermissionManager
	bans                     map[string]ban
	commandPermissionManager permissions.PermissionManager
	ctx                      context.Context
	cancelFunc               context.CancelFunc
}

func (b *Bot) UserHasPermission(user *messages.User, permission permissions.Permission) bool {
	perm, err := b.userPermissionManager.GetPermission(user.GetId())
	if err != nil {
		return false
	}
	return perm.Has(permission)
}

func (b *Bot) OnlineUsers() map[string]messages.User {
	users := map[string]messages.User{}
	for _, user := range b.connectorRelay.GetUsers() {
		users[user.GetNick()] = messages.User{
			Nick:  user.GetNick(),
			Id:    user.GetId(),
			Mod:   user.GetMod(),
			Admin: user.GetAdmin(),
		}
	}
	return users
}

func (b *Bot) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (b *Bot) Done() <-chan struct{} {
	return b.ctx.Done()
}

func (b *Bot) Err() error {
	return b.ctx.Err()
}

func (b *Bot) Value(key interface{}) interface{} {
	return b.ctx.Value(key)
}

func NewBot(config bot.Config, userPermissionManager permissions.PermissionManager, commandPermissionManager permissions.PermissionManager) *Bot {
	return &Bot{bans: make(map[string]ban), commands: make(map[string]command.Command), config: config, commandPermissionManager: commandPermissionManager, userPermissionManager: userPermissionManager, connectorRelay: relay.GetConnectorRelay(config)}
}

func (b *Bot) BotUser() *messages.User {
	return b.botUser
}

func (b *Bot) ApiKeys() map[string]string {
	return b.config.Bot.ApiKeys
}

func (b *Bot) AddCommand(command command.Command) {
	if _, exists := b.commands[command.Name()]; exists {
		return
	}
	b.commands[command.Name()] = command
}

func (b Bot) isCommandDisabled(command command.Command) bool {
	isDisabled, exists := b.config.Bot.Commands.Disabled[command.Name()]
	if !exists {
		isDisabled = false
	}
	return isDisabled
}

func (b *Bot) getCommandList() []*messages.Command {
	var commands []*messages.Command
	for _, cmd := range b.commands {
		if b.isCommandDisabled(cmd) {
			continue
		}
		commands = append(commands, &messages.Command{
			Name:    cmd.Name(),
			Aliases: cmd.Aliases(),
			Usage:   fmt.Sprintf("%s%s <args>", b.config.Bot.Trigger, cmd.Name()),
		})
	}
	return commands
}

func (b *Bot) Start() error {
	b.ctx, b.cancelFunc = context.WithCancel(context.Background())
	var err error
	b.AddCommand(&builtinCommand{
		NoOpCommand: command.NoOpCommand{},
		name:        "ban",
		execute:     b.ban,
	})
	b.AddCommand(&builtinCommand{
		NoOpCommand: command.NoOpCommand{},
		name:        "verify",
		execute:     b.verify,
	})
	for _, cmd := range Commands {
		if b.isCommandDisabled(cmd) {
			continue
		}
		err := cmd.Init(b)
		if err != nil {
			log.Printf("couldn't init %s\n", cmd.Name())
			b.config.Bot.Commands.Disabled[cmd.Name()] = true
		}
		b.AddCommand(cmd)
	}
	confirmation, err := b.connectorRelay.Connect(&messages.RegistrationPacket{
		Trigger:  "!",
		Commands: b.getCommandList(),
	})
	if err != nil {
		return err
	}
	b.botUser = confirmation
	go func() {
		<-b.connectorRelay.Done()
		b.cancelFunc()
	}()
	go func() {
		var relayError error
		var message relay.ConnectorMessage
		var packets []*messages.BotPacket
		for ; relayError == nil; relayError = b.connectorRelay.RecvMsg(&message) {
			relayError := relayError
			go func() {
				var err error
				switch message.Message.(type) {
				case *messages.CommandPacket:
					packets, err = b.parseCommandPacket(message.Message.(*messages.CommandPacket))
				case *messages.MessagePacket:
					packets, err = b.parseMessagePacket(message.Message.(*messages.MessagePacket))
				}
				if err != nil {
					return
				}
				for _, packet := range packets {
					if packet == nil {
						return
					}
					relayError = b.connectorRelay.Send(packet)
					if relayError != nil {
						log.Println("error", relayError.Error())
						b.cancelFunc()
						return
					}
				}
			}()
		}
	}()
	return nil
}

func (b *Bot) parseCommandPacket(packet *messages.CommandPacket) ([]*messages.BotPacket, error) {
	log.Println("Parsing", packet.GetCommand(), packet.GetArgs())
	if b.isBanned(packet.GetUser()) {
		return nil, nil
	}
	var c command.Command
	var ok bool
	if c, ok = b.commands[packet.GetCommand()]; ok {
		for _, cmd := range b.commands {
			for _, alias := range cmd.Aliases() {
				if packet.GetCommand() == alias {
					c = cmd
					break
				}
			}
		}
	}
	if c == nil {
		return nil, nil
	}
	log.Println("Command", c.Name())
	if b.isCommandDisabled(c) {
		return nil, nil
	}
	log.Println("is not disabled")
	if !b.isAllowed(c.Name(), packet.GetUser()) {
		return nil, nil
	}
	log.Println("is allowed")
	return c.Execute(packet)
}

func (b *Bot) parseMessagePacket(packet *messages.MessagePacket) ([]*messages.BotPacket, error) {
	if b.isBanned(packet.User) {
		return nil, nil
	}
	var packets []*messages.BotPacket
	for _, cmd := range b.commands {
		if b.isCommandDisabled(cmd) {
			continue
		}
		if cmd.IgnoreSelf() {
			if packet.GetUser().GetNick() == b.botUser.GetNick() {
				continue
			}
		}
		response, err := cmd.OnChat(packet)
		if err != nil {
			continue
		}
		packets = append(packets, response...)
	}
	return packets, nil
}

func (b *Bot) isAllowed(command string, user *messages.User) bool {
	uPermission, err := b.userPermissionManager.GetPermission(user.Id)
	if err != nil {
		return false
	}
	cPermission, err := b.commandPermissionManager.GetPermission(command)
	if err != nil {
		return false
	}
	return uPermission.Has(cPermission)
}

func (b *Bot) isBanned(user *messages.User) bool {
	bn, exists := b.bans[user.Nick]
	if !exists {
		return false
	}
	if bn.duration < 0 || bn.start.Add(bn.duration).After(time.Now()) {
		return true
	}
	delete(b.bans, user.Nick)
	return false
}
