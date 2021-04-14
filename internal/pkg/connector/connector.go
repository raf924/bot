package connector

import (
	"context"
	"fmt"
	"github.com/raf924/bot/pkg/command"
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/queue"
	"github.com/raf924/bot/pkg/relay/connection"
	"github.com/raf924/bot/pkg/relay/server"
	"github.com/raf924/bot/pkg/users"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"sort"
	"strings"
	"time"
)

var helpCommand = &messages.Command{
	Name:    "help",
	Aliases: []string{"h"},
	Usage:   "help",
}

type Connector struct {
	config          connector.Config
	connectionRelay connection.Relay
	relayServer     server.RelayServer
	context         context.Context
	cancelFunc      context.CancelFunc
	users           *users.UserList
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

func (c *Connector) getCommandOr(mP *messages.MessagePacket) proto.Message {
	log.Printf("Message from %s: %s\n", mP.GetUser().GetNick(), mP.GetMessage())
	if len(c.config.Trigger) == 0 {
		return mP
	}
	if !strings.HasPrefix(mP.GetMessage(), c.config.Trigger) {
		return mP
	}
	argString := strings.TrimPrefix(mP.GetMessage(), c.config.Trigger)
	args := strings.Split(argString, " ")
	if len(args) == 0 || len(args[0]) == 0 {
		return mP
	}
	possibleCommand := args[0]
	if command.Is(possibleCommand, helpCommand) {
		var names []string
		for _, cmd := range c.relayServer.Commands() {
			names = append(names, fmt.Sprintf("%s%s", c.config.Trigger, cmd.GetName()))
			for _, alias := range append(cmd.Aliases) {
				names = append(names, fmt.Sprintf("%s%s (%s)", c.config.Trigger, alias, cmd.GetName()))
			}
		}
		sort.Strings(names)
		err := c.sendToConnection(connection.ChatMessage{
			Message:   strings.Join(names, ", "),
			Recipient: mP.GetUser().GetNick(),
			Private:   mP.GetPrivate(),
		})
		if err != nil {
			panic(err)
		}
		return nil
	}
	cmd := command.Find(c.relayServer.Commands(), possibleCommand)
	if cmd == nil {
		return mP
	}
	argString = strings.TrimSpace(strings.TrimPrefix(argString, possibleCommand))
	return &messages.CommandPacket{
		Timestamp: mP.GetTimestamp(),
		Command:   cmd.GetName(),
		Args:      args[1:],
		ArgString: argString,
		User:      mP.GetUser(),
		Private:   mP.GetPrivate(),
	}
}

func (c *Connector) Start() error {
	c.context, c.cancelFunc = context.WithCancel(context.Background())
	c.connectionRelay.OnUserJoin(func(user *messages.User, timestamp int64) {
		c.users = c.connectionRelay.GetUsers()
		err := c.relayServer.Send(&messages.UserPacket{
			Timestamp: timestamppb.New(time.Unix(timestamp, 0)),
			User:      user,
			Event:     messages.UserEvent_JOINED,
		})
		if err != nil {
			log.Println(err)
		}
	})
	c.connectionRelay.OnUserLeft(func(user *messages.User, timestamp int64) {
		c.users = c.connectionRelay.GetUsers()
		err := c.relayServer.Send(&messages.UserPacket{
			Timestamp: timestamppb.New(time.Unix(timestamp, 0)),
			User:      user,
			Event:     messages.UserEvent_LEFT,
		})
		if err != nil {
			log.Println(err)
		}
	})
	err := c.connectionRelay.Connect(c.config.Name)
	if err != nil {
		return err
	}
	u := c.connectionRelay.GetUsers().Find(c.config.Name)
	if u == nil {
		return fmt.Errorf("couldn't find connector among users")
	}
	err = c.relayServer.Start(u, c.connectionRelay.GetUsers().All(), c.config.Trigger)
	if err != nil {
		return err
	}
	go func() {
		for {
			mP, err := c.receiveFromConnection()
			if err != nil {
				c.cancelFunc()
				return
			}
			m := c.getCommandOr(mP)
			if m == nil {
				continue
			}
			go func() {
				err := c.relayServer.Send(m)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}()
	go func() {
		for {
			packet, err := c.receiveFromRelayServer()
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
			err = c.sendToConnection(connection.ChatMessage{
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

func (c *Connector) receiveFromRelayServer() (*messages.BotPacket, error) {
	return c.relayServer.Recv()
}

func (c *Connector) receiveFromConnection() (*messages.MessagePacket, error) {
	return c.connectionRelay.Recv()
}

func (c *Connector) sendToServer(m proto.Message) error {
	return c.relayServer.Send(m)
}

func (c *Connector) sendToConnection(m connection.Message) error {
	return c.connectionRelay.Send(m)
}

func NewConnector(config connector.Config, connection connection.Relay, bot server.RelayServer, connectionExchange, botExchange *queue.Exchange) *Connector {
	return &Connector{
		config:          config,
		connectionRelay: connection,
		relayServer:     bot,
	}
}
