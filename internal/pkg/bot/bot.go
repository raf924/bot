package bot

import (
	"context"
	"fmt"
	_ "github.com/raf924/bot/v2/internal/pkg/bot/permissions"
	"github.com/raf924/bot/v2/pkg"
	"github.com/raf924/bot/v2/pkg/bot/permissions"
	"github.com/raf924/bot/v2/pkg/config/bot"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
	"github.com/raf924/connector-sdk/rpc"
	"github.com/raf924/connector-sdk/storage"
	"log"
	"time"
)

var _ pkg.Runnable = (*Bot)(nil)

type ban struct {
	Start    time.Time     `json:"Start"`
	Duration time.Duration `json:"Duration"`
}

type Bot struct {
	connectorRelay           rpc.DispatcherRelay
	users                    domain.UserList
	loadedCommands           map[string]command.Command
	commands                 command.List
	config                   bot.Config
	botUser                  *domain.User
	bans                     map[string]ban
	userPermissionManager    permissions.PermissionManager
	commandPermissionManager permissions.PermissionManager
	ctx                      context.Context
	cancelFunc               func(err error)
	banStorage               storage.Storage
	trigger                  string
}

func NewBot(
	config bot.Config,
	userPermissionManager permissions.PermissionManager,
	commandPermissionManager permissions.PermissionManager,
	relay rpc.DispatcherRelay,
	commands command.List,
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
		banStorage:               banStorage,
	}
}

func (b *Bot) Trigger() string {
	return b.trigger
}

func (b *Bot) UserHasPermission(user *domain.User, permission domain.Permission) bool {
	perm, err := b.userPermissionManager.GetPermission(user.Id())
	if err != nil {
		return false
	}
	return perm.Has(permission)
}

func (b *Bot) OnlineUsers() domain.UserList {
	return domain.ImmutableUserList(b.users)
}

func (b *Bot) Done() <-chan struct{} {
	return b.ctx.Done()
}

func (b *Bot) Err() error {
	return b.ctx.Err()
}

func (b *Bot) BotUser() *domain.User {
	return b.botUser
}

func (b *Bot) ApiKeys() map[string]string {
	return b.config.ApiKeys
}

func (b *Bot) Start(ctx context.Context) error {
	b.ctx, b.cancelFunc = pkg.Errorable(ctx)
	go func() {
		select {
		case <-b.connectorRelay.Done():
			b.cancelFunc(fmt.Errorf("connector error: %w", b.connectorRelay.Err()))
		}
	}()
	b.loadBans()
	b.initCommands()
	var err error
	confirmation, err := b.connectorRelay.Connect(domain.NewRegistrationMessage(b.getCommandList()))
	if err != nil {
		return fmt.Errorf("cannot connect to server: %w", err)
	}
	b.botUser = confirmation.CurrentUser()
	b.users = confirmation.Users()
	b.trigger = confirmation.Trigger()
	commandHandler := CommandHandler{
		commands:       domain.ImmutableCommandList(domain.NewCommandList(b.getCommandList()...)),
		loadedCommands: b.loadedCommands,
		botUser:        b.botUser,
		commandCallback: func(messages []*domain.ClientMessage, err error) error {
			if b.ctx.Err() != nil {
				return fmt.Errorf("bot is down: %v", b.ctx.Err())
			}
			if err != nil {
				log.Println("error running command", err)
				return nil
			}
			for _, message := range messages {
				if b.ctx.Err() != nil {
					return fmt.Errorf("bot is down: %v", b.ctx.Err())
				}
				go func(message *domain.ClientMessage) {
					err := b.connectorRelay.Send(message)
					if err != nil {
						b.cancelFunc(err)
					}
				}(message)
			}
			return nil
		},
		userPermissionManager:    b.userPermissionManager,
		commandPermissionManager: b.commandPermissionManager,
	}
	go func() {
		for b.ctx.Err() == nil {
			packet, err := b.connectorRelay.Recv()
			if err != nil {
				b.cancelFunc(err)
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
			go func() {
				senderIsBanned := false
				if packet, ok := packet.(FromUser); ok {
					senderIsBanned = b.isBanned(packet.Sender())
				}
				if err = commandHandler.PassServerMessage(packet, senderIsBanned); err != nil {
					b.cancelFunc(err)
					return
				}
			}()
		}
	}()
	return nil
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
	b.commands.Add(&builtinCommand{
		NoOpCommand: command.NoOpCommand{},
		name:        "ban",
		execute:     b.ban,
	})
	b.commands.Add(&builtinCommand{
		NoOpCommand: command.NoOpCommand{},
		name:        "verify",
		execute:     b.verify,
	})
	b.commands.Range(func(command command.Command) bool {
		if b.isCommandDisabled(command) {
			return true
		}
		err := command.Init(b)
		if err != nil {
			log.Printf("couldn't init %s\n", command.Name())
			b.disable(command)
		}
		b.AddCommand(command)
		return true
	})
}

type FromUser interface {
	Sender() *domain.User
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
