package bot

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/raf924/bot/internal/pkg/bot/permissions"
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/bot/permissions"
	"github.com/raf924/bot/pkg/config/bot"
	"github.com/raf924/bot/pkg/queue"
	"github.com/raf924/bot/pkg/relay/client"
	messages "github.com/raf924/connector-api/pkg/gen"
	"log"
	"os"
	"sync"
	"time"
)

var Commands []command.Command

type ban struct {
	Start    time.Time     `json:"Start"`
	Duration time.Duration `json:"Duration"`
}

type Bot struct {
	connectorRelay           client.RelayClient
	loadedCommands           map[string]command.Command
	commands                 []command.Command
	config                   bot.Config
	botUser                  *messages.User
	userPermissionManager    permissions.PermissionManager
	bans                     map[string]ban
	commandPermissionManager permissions.PermissionManager
	ctx                      context.Context
	cancelFunc               context.CancelFunc
	commandQueue             queue.Queue
	banStorageMutex          *sync.Mutex
}

func (b *Bot) Trigger() string {
	return b.config.Trigger
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

func NewBot(
	config bot.Config,
	userPermissionManager permissions.PermissionManager,
	commandPermissionManager permissions.PermissionManager,
	relay client.RelayClient,
	commands ...command.Command,
) *Bot {
	return &Bot{
		bans:                     make(map[string]ban),
		loadedCommands:           make(map[string]command.Command),
		commands:                 commands,
		config:                   config,
		commandPermissionManager: commandPermissionManager,
		userPermissionManager:    userPermissionManager,
		connectorRelay:           relay,
		commandQueue:             queue.NewQueue(),
	}
}

func (b *Bot) BotUser() *messages.User {
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

func (b *Bot) getCommandList() []*messages.Command {
	var commands []*messages.Command
	for _, cmd := range b.loadedCommands {
		if b.isCommandDisabled(cmd) {
			continue
		}
		commands = append(commands, &messages.Command{
			Name:    cmd.Name(),
			Aliases: cmd.Aliases(),
			Usage:   fmt.Sprintf("%s%s <args>", b.config.Trigger, cmd.Name()),
		})
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
		go func(cmd command.Command) {
			err := b.relayBotPackets(cmd, consumer)
			if err != nil {
				log.Println("command", cmd.Name(), "disabled:", err.Error())
				b.disable(cmd)
			}
			consumer.Cancel()
		}(cmd)
	}
}

func (b *Bot) Start() error {
	b.banStorageMutex = &sync.Mutex{}
	b.ctx, b.cancelFunc = context.WithCancel(context.Background())
	b.loadBans()
	var err error
	b.initCommands()
	confirmation, err := b.connectorRelay.Connect(&messages.RegistrationPacket{
		Trigger:  b.config.Trigger,
		Commands: b.getCommandList(),
	})
	if err != nil {
		return err
	}
	b.botUser = confirmation
	b.connectorRelay.OnUserLeft(func(user *messages.User, timestamp int64) {

	})
	b.connectorRelay.OnUserJoin(func(user *messages.User, timestamp int64) {

	})
	go func() {
		<-b.connectorRelay.Done()
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
				panic(err)
			}
			if err = commandProducer.Produce(packet); err != nil {
				panic(err)
			}
		}
	}()
	return nil
}

type FromUser interface {
	GetUser() *messages.User
}

func (b *Bot) relayBotPackets(cmd command.Command, commandConsumer *queue.Consumer) error {
	for {
		m, err := commandConsumer.Consume()
		if err != nil {
			return err
		}
		var packets []*messages.BotPacket
		var sender = m.(FromUser).GetUser()
		if sender.GetNick() == b.botUser.GetNick() && cmd.IgnoreSelf() {
			log.Println("Command", cmd.Name(), "ignored message from self")
			continue
		}
		if !b.isAllowed(cmd.Name(), sender) {
			log.Println(sender.GetNick(), "is not allowed to use", cmd.Name())
			continue
		}
		switch m.(type) {
		case *messages.MessagePacket:
			message := m.(*messages.MessagePacket)
			packets, err = cmd.OnChat(message)
			if err != nil {
				log.Println("Command", cmd.Name(), "OnChat error:", err.Error())
			}
		case *messages.CommandPacket:
			message := m.(*messages.CommandPacket)
			if b.isBanned(message.GetUser()) {
				continue
			}
			if !command.Is(message.GetCommand(), cmd) {
				packets, err = cmd.OnChat(&messages.MessagePacket{
					Timestamp: message.GetTimestamp(),
					Message:   fmt.Sprintf("%s %s", message.GetCommand(), message.GetArgString()),
					User:      message.GetUser(),
					Private:   message.GetPrivate(),
				})
				if err != nil {
					log.Println("Command", cmd.Name(), "Execute error:", err.Error())
				}
				break
			}
			packets, err = cmd.Execute(message)
			if err != nil {
				log.Println("Command", cmd.Name(), "Execute error:", err.Error())
			}
		case *messages.UserPacket:
			message := m.(*messages.UserPacket)
			packets, err = cmd.OnUserEvent(message)
			if err != nil {
				log.Println("Command", cmd.Name(), "OnUserEvent error:", err.Error())
			}
		}
		for _, packet := range packets {
			err := b.connectorRelay.Send(packet)
			if err != nil {
				return err
			}
		}
	}
}

func (b *Bot) isAllowed(command string, user *messages.User) bool {
	uPermission, err := b.userPermissionManager.GetPermission(user.GetId())
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
	if bn.Duration < 0 || bn.Start.Add(bn.Duration).After(time.Now()) {
		return true
	}
	delete(b.bans, user.Nick)
	return false
}

func (b *Bot) loadBans() {
	banStorageLocation := b.config.ApiKeys["banStorageLocation"]
	f, err := os.Open(banStorageLocation)
	if err != nil {
		log.Println("could not load bans: ", err)
		return
	}
	_ = json.NewDecoder(f).Decode(&b.bans)
	_ = f.Close()
}

func (b *Bot) saveBans() {
	go func() {
		b.banStorageMutex.Lock()
		defer b.banStorageMutex.Unlock()
		banStorageLocation := b.config.ApiKeys["banStorageLocation"]
		file, err := os.OpenFile(banStorageLocation, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			log.Println("error opening ban file:", err.Error())
			return
		}
		err = json.NewEncoder(file).Encode(b.bans)
		_ = file.Close()
		if err != nil {
			log.Println("error writing to ban file:", err.Error())
			return
		}
	}()

}
