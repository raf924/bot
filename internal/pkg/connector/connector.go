package connector

import (
	"context"
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/relay"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"sort"
	"strings"
	"time"
)

type Connector struct {
	config          connector.Config
	connectionRelay relay.ConnectionRelay
	botRelay        relay.BotRelay
	context         context.Context
	cancelFunc      context.CancelFunc
}

func (c *Connector) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (c *Connector) Done() <-chan struct{} {
	return c.context.Done()
}

func (c *Connector) Err() error {
	return c.context.Err()
}

func (c *Connector) Value(key interface{}) interface{} {
	return c.context.Value(key)
}

func (c *Connector) Start() error {
	c.context, c.cancelFunc = context.WithCancel(context.Background())
	c.connectionRelay.OnUserJoin(func(user *messages.User, timestamp int64) {
		err := c.botRelay.PassEvent(&messages.UserPacket{
			Timestamp: timestamppb.New(time.Unix(timestamp, 0)),
			User:      user,
			Event:     messages.UserEvent_JOINED,
		})
		if err != nil {
			log.Println(err)
		}
	})
	c.connectionRelay.OnUserLeft(func(user *messages.User, timestamp int64) {
		err := c.botRelay.PassEvent(&messages.UserPacket{
			Timestamp: timestamppb.New(time.Unix(timestamp, 0)),
			User:      user,
			Event:     messages.UserEvent_LEFT,
		})
		if err != nil {
			log.Println(err)
		}
	})
	err := c.connectionRelay.Connect(c.config.Connector.Name)
	if err != nil {
		return err
	}
	//TODO: something better
	for _, user := range c.connectionRelay.GetUsers() {
		if user.GetNick() == c.config.Connector.Name {
			err = c.botRelay.Start(user, c.connectionRelay.GetUsers())
			break
		}
	}
	if err != nil {
		return err
	}
	go func() {
		for {
			mP := &messages.MessagePacket{}
			if err := c.connectionRelay.RecvMsg(mP); err != nil {
				c.cancelFunc()
				return
			}
			log.Printf("Message from %s: %s\n", mP.GetUser().GetNick(), mP.GetMessage())
			if c.botRelay.Trigger() == "" {
				continue
			}
			var actualCommand = ""
			if strings.HasPrefix(mP.GetMessage(), c.botRelay.Trigger()) {
				argString := strings.TrimPrefix(mP.Message, c.botRelay.Trigger())
				args := strings.Split(argString, " ")
				if len(args) > 0 && len(args[0]) > 0 {
					command := args[0]
					argString := strings.TrimSpace(strings.TrimPrefix(argString, command))
					log.Println("Command", command)
					for _, cmd := range c.botRelay.Commands() {
						if command == cmd.GetName() {
							log.Println("Found command", command, cmd.GetName())
							actualCommand = cmd.GetName()
							break
						}
						aliases := cmd.GetAliases()
						if len(aliases) == 0 {
							continue
						}
						sort.Strings(aliases)
						index := sort.SearchStrings(aliases, command)
						if index < len(aliases) && aliases[index] == command {
							log.Println("Found alias")
							actualCommand = cmd.GetName()
							break
						}
					}
					log.Println("Passing", actualCommand)
					if len(actualCommand) > 0 {
						err := c.botRelay.PassCommand(&messages.CommandPacket{
							Timestamp: mP.GetTimestamp(),
							Command:   actualCommand,
							Args:      args[1:],
							ArgString: argString,
							User:      mP.GetUser(),
							Private:   mP.GetPrivate(),
						})
						if err != nil {
							log.Println(err)
						}
						continue
					}
				}
			}
			if err := c.botRelay.PassMessage(mP); err != nil {
				log.Println(err)
			}
		}
	}()
	go func() {
		<-c.botRelay.Ready()
		for {
			var packet = messages.BotPacket{}
			err := c.botRelay.RecvMsg(&packet)
			if err != nil {
				log.Println(err)
				continue
			}
			if packet.Timestamp == nil {
				continue
			}
			recipient := ""
			if packet.Recipient != nil {
				recipient = packet.Recipient.Nick
			}
			err = c.connectionRelay.Send(relay.ChatMessage{
				Message:   packet.Message,
				Recipient: recipient,
				Private:   packet.Private,
			})
			if err != nil {
				c.cancelFunc()
				return
			}
		}
	}()
	return nil
}

func NewConnector(config connector.Config) *Connector {
	return &Connector{
		config:          config,
		connectionRelay: relay.GetConnectionRelay(config),
		botRelay:        relay.GetBotRelay(config),
	}
}
