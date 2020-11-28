package connector

import (
	"context"
	"github.com/raf924/bot/api/messages"
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/relay"
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
	//TODO: somthing better
	for _, user := range c.connectionRelay.GetUsers() {
		if user.Nick == c.config.Connector.Name {
			err = c.botRelay.Start(user)
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
			log.Printf("%s: %s\n", mP.User.Nick, mP.Message)
			if c.botRelay.Trigger() == "" {
				continue
			}
			var isCommand = false
			if strings.HasPrefix(mP.Message, c.botRelay.Trigger()) {
				args := strings.Split(strings.TrimPrefix(mP.Message, c.botRelay.Trigger()), " ")
				if len(args) > 0 && len(args[0]) > 0 {
					command := args[0]
					for _, cmd := range c.botRelay.Commands() {
						if command == cmd.Name {
							isCommand = true
							break
						}
						if len(cmd.Aliases) == 0 {
							continue
						}
						sort.Strings(cmd.Aliases)
						index := sort.SearchStrings(cmd.Aliases, command)
						if index == len(cmd.Aliases) || cmd.Aliases[index] == command {
							isCommand = true
							break
						}
					}
					if isCommand {
						err := c.botRelay.PassCommand(&messages.CommandPacket{
							Timestamp: mP.Timestamp,
							Command:   args[0],
							Args:      args[1:],
							User:      mP.User,
							Private:   mP.Private,
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
			err = c.connectionRelay.Send(relay.ChatMessage{
				Message:   packet.Message,
				Recipient: packet.Recipient.Nick,
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
