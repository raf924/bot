package bot

import (
	"context"
	"fmt"
	_ "github.com/raf924/bot/internal/pkg/bot/permissions"
	"github.com/raf924/bot/pkg"
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/bot/permissions"
	"github.com/raf924/bot/pkg/config/bot"
	"github.com/raf924/bot/pkg/domain"
	"github.com/raf924/bot/pkg/queue"
	"github.com/raf924/bot/pkg/relay/client"
	"github.com/raf924/bot/pkg/storage"
	"log"
	"time"
)

var _ pkg.Runnable = (*Bot)(nil)
var _ context.Context = (*Bot)(nil)

var Commands []command.Command

type ban struct {
	Start    time.Time     `json:"Start"`
	Duration time.Duration `json:"Duration"`
}

type Bot struct {
	connectorRelay           client.RelayClient
	users                    domain.UserList
	loadedCommands           map[string]command.Command
	commands                 []command.Command
	config                   bot.Config
	botUser                  *domain.User
	userPermissionManager    permissions.PermissionManager
	bans                     map[string]ban
	commandPermissionManager permissions.PermissionManager
	ctx                      context.Context
	cancelFunc               context.CancelFunc
	commandQueue             queue.Queue
	banStorage               storage.Storage
	errorChan                chan error
	trigger                  string
}

func (b *Bot) Trigger() string {
	return b.trigger
}

func (b *Bot) UserHasPermission(user *domain.User, permission permissions.Permission) bool {
	perm, err := b.userPermissionManager.GetPermission(user.Id())
	if err != nil {
		return false
	}
	return perm.Has(permission)
}

func (b *Bot) OnlineUsers() domain.UserList {
	users := b.users
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

func NewBot(
	config bot.Config,
	userPermissionManager permissions.PermissionManager,
	commandPermissionManager permissions.PermissionManager,
	relay client.RelayClient,
	commands ...command.Command,
) *Bot {
	banStorage, err := storage.NewFileStorage(config.ApiKeys["banStorageLocation"])
	if err != nil {
		log.Println(err)
		banStorage = storage.NewNoOpStorage()
	}
	return &Bot{
		users:                    domain.NewUserList(),
		bans:                     make(map[string]ban),
		loadedCommands:           make(map[string]command.Command),
		commands:                 commands,
		config:                   config,
		commandPermissionManager: commandPermissionManager,
		userPermissionManager:    userPermissionManager,
		connectorRelay:           relay,
		commandQueue:             queue.NewQueue(),
		banStorage:               banStorage,
	}
}

func (b *Bot) BotUser() *domain.User {
	return b.botUser
}

func (b *Bot) ApiKeys() map[string]string {
	return b.config.ApiKeys
}

func (b *Bot) AddCommand(command command.Command) {
	if _, exists := b.loadedCommands[command.Name()]; exists {
		return
	}
	b.loadedCommands[command.Name()] = command
}

func (b *Bot) isCommandDisabled(command command.Command) bool {
	if b.config.Commands.Disabled == nil {
		return false
	}
	isDisabled, exists := b.config.Commands.Disabled[command.Name()]
	if !exists {
		isDisabled = false
	}
	return isDisabled
}

func (b *Bot) getCommandList() []*domain.Command {
	var commands []*domain.Command
	for _, cmd := range b.loadedCommands {
		if b.isCommandDisabled(cmd) {
			continue
		}
		commands = append(commands, domain.NewCommand(cmd.Name(), cmd.Aliases(), fmt.Sprintf("%s%s <args>", b.trigger, cmd.Name())))
	}
	return commands
}

func (b *Bot) disable(cmd command.Command) {
	b.config.Commands.Disabled[cmd.Name()] = true
}

func (b *Bot) initCommands() {
	b.commands = append(b.commands, &builtinCommand{
		NoOpCommand: command.NoOpCommand{},
		name:        "ban",
		execute:     b.ban,
	}, &builtinCommand{
		NoOpCommand: command.NoOpCommand{},
		name:        "verify",
		execute:     b.verify,
	})
	for _, cmd := range b.commands {
		if b.isCommandDisabled(cmd) {
			continue
		}
		err := cmd.Init(b)
		if err != nil {
			log.Printf("couldn't init %s\n", cmd.Name())
			b.disable(cmd)
		}
		b.AddCommand(cmd)
		consumer, err := b.commandQueue.NewConsumer()
		if err != nil {
			log.Printf("couldn't init %s\n", cmd.Name())
			b.disable(cmd)
			continue
		}
		cmdContext, cmdCancel := context.WithCancel(b.ctx)
		go func(consumer *queue.Consumer) {
			<-cmdContext.Done()
			consumer.Cancel()
		}(consumer)
		go func(cmd command.Command) {
			err := b.relayBotPackets(cmd, consumer)
			if err != nil {
				log.Println("command", cmd.Name(), "disabled:", err.Error())
				b.disable(cmd)
			}
			cmdCancel()
		}(cmd)
	}
}

func (b *Bot) Start() error {
	b.ctx, b.cancelFunc = context.WithCancel(context.Background())
	b.loadBans()
	var err error
	b.initCommands()
	confirmation, err := b.connectorRelay.Connect(domain.NewRegistrationMessage(b.getCommandList()))
	if err != nil {
		return fmt.Errorf("cannot connect to server: %v", err)
	}
	b.botUser = confirmation.CurrentUser()
	b.users = confirmation.Users()
	b.trigger = confirmation.Trigger()
	go func() {
		err := <-b.errorChan
		log.Println(err)
		b.cancelFunc()
	}()
	commandProducer, err := b.commandQueue.NewProducer()
	if err != nil {
		return err
	}
	go func() {
		for {
			packet, err := b.connectorRelay.Recv()
			if err != nil {
				b.errorChan <- err
				return
			}
			switch packet := packet.(type) {
			case *domain.UserEvent:
				switch packet.EventType() {
				case domain.UserJoined:
					b.users.Add(packet.User())
				case domain.UserLeft:
					b.users.Remove(packet.User())
				}
			}
			if err = commandProducer.Produce(packet); err != nil {
				b.errorChan <- err
				return
			}
		}
	}()
	return nil
}

type FromUser interface {
	Sender() *domain.User
}

func (b *Bot) relayBotPackets(cmd command.Command, commandConsumer *queue.Consumer) error {
	cmdAliases := map[string]struct{}{cmd.Name(): {}}
	for _, s := range cmd.Aliases() {
		cmdAliases[s] = struct{}{}
	}
	for {
		m, err := commandConsumer.Consume()
		if err != nil {
			return err
		}
		var packets []*domain.ClientMessage
		var sender = m.(FromUser).Sender()
		if sender.Nick() == b.botUser.Nick() && cmd.IgnoreSelf() {
			log.Println("Command", cmd.Name(), "ignored message from self")
			continue
		}
		if !b.isAllowed(cmd.Name(), sender) {
			log.Println(sender.Nick(), "is not allowed to use", cmd.Name())
			continue
		}
		switch message := m.(type) {
		case *domain.ChatMessage:
			packets, err = cmd.OnChat(message)
			if err != nil {
				log.Println("Command", cmd.Name(), "OnChat error:", err.Error())
			}
		case *domain.CommandMessage:
			if b.isBanned(message.Sender()) {
				continue
			}
			if _, isCmd := cmdAliases[message.Command()]; !isCmd {
				packets, err = cmd.OnChat(message.ToChatMessage())
				if err != nil {
					log.Println("Command", cmd.Name(), "OnChat error:", err.Error())
				}
			} else {
				packets, err = cmd.Execute(message)
				if err != nil {
					log.Println("Command", cmd.Name(), "Execute error:", err.Error())
				}
			}
		case *domain.UserEvent:
			packets, err = cmd.OnUserEvent(message)
			if err != nil {
				log.Println("Command", cmd.Name(), "OnUserEvent error:", err.Error())
			}
		}
		if err != nil {
			continue
		}
		for _, packet := range packets {
			err := b.connectorRelay.Send(packet)
			if err != nil {
				return err
			}
		}
	}
}

func (b *Bot) isAllowed(command string, user *domain.User) bool {
	uPermission, err := b.userPermissionManager.GetPermission(user.Id())
	if err != nil {
		return false
	}
	cPermission, err := b.commandPermissionManager.GetPermission(command)
	if err != nil {
		return false
	}
	return uPermission.Has(cPermission)
}

func (b *Bot) isBanned(user *domain.User) bool {
	bn, exists := b.bans[user.Nick()]
	if !exists {
		return false
	}
	if bn.Duration < 0 || bn.Start.Add(bn.Duration).After(time.Now()) {
		return true
	}
	delete(b.bans, user.Nick())
	return false
}

func (b *Bot) loadBans() {
	err := b.banStorage.Load(&b.bans)
	if err != nil {
		log.Println("could not load bans: ", err)
		return
	}
}

func (b *Bot) saveBans() {
	b.banStorage.Save(b.bans)
}
