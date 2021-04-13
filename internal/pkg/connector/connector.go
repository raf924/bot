package connector

import (
	"context"
	"github.com/raf924/bot/pkg/command"
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/queue"
	"github.com/raf924/bot/pkg/relay/connection"
	"github.com/raf924/bot/pkg/relay/server"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"strings"
	"time"
)

type Connector struct {
	config             connector.Config
	connectionRelay    connection.ConnectionRelay
	botRelay           server.RelayServer
	context            context.Context
	cancelFunc         context.CancelFunc
	connectionExchange *queue.Exchange
	serverExchange     *queue.Exchange
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
	if len(c.botRelay.Trigger()) == 0 {
		return mP
	}
	if !strings.HasPrefix(mP.GetMessage(), c.botRelay.Trigger()) {
		return mP
	}
	argString := strings.TrimPrefix(mP.GetMessage(), c.botRelay.Trigger())
	args := strings.Split(argString, " ")
	if len(args) == 0 || len(args[0]) == 0 {
		return mP
	}
	possibleCommand := args[0]
	cmd := command.Find(c.botRelay.Commands(), possibleCommand)
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
		err := c.sendToServer(&messages.UserPacket{
			Timestamp: timestamppb.New(time.Unix(timestamp, 0)),
			User:      user,
			Event:     messages.UserEvent_JOINED,
		})
		if err != nil {
			log.Println(err)
		}
	})
	c.connectionRelay.OnUserLeft(func(user *messages.User, timestamp int64) {
		err := c.sendToServer(&messages.UserPacket{
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
	//TODO: something better
	for _, user := range c.connectionRelay.GetUsers() {
		if user.GetNick() == c.config.Name {
			err = c.botRelay.Start(user, c.connectionRelay.GetUsers(), c.config.Trigger)
			break
		}
	}
	if err != nil {
		return err
	}
	go func() {
		for {
			mP := &messages.MessagePacket{}
			mP, err := c.receiveFromConnection()
			if err != nil {
				c.cancelFunc()
				return
			}
			m := c.getCommandOr(mP)
			err = c.sendToServer(m)
			if err != nil {
				log.Println(err)
			}
		}
	}()
	go func() {
		for {
			packet, err := c.receiveFromBot()
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

func (c *Connector) receiveFromBot() (*messages.BotPacket, error) {
	v, err := c.serverExchange.Consume()
	return v.(*messages.BotPacket), err
}

func (c *Connector) receiveFromConnection() (*messages.MessagePacket, error) {
	v, err := c.connectionExchange.Consume()
	return v.(*messages.MessagePacket), err
}

func (c *Connector) sendToServer(m proto.Message) error {
	return c.serverExchange.Produce(m)
}

func (c *Connector) sendToConnection(m connection.Message) error {
	return c.connectionExchange.Produce(m)
}

func NewConnector(config connector.Config, connection connection.ConnectionRelay, bot server.RelayServer, connectionExchange, botExchange *queue.Exchange) *Connector {
	return &Connector{
		config:             config,
		connectionRelay:    connection,
		botRelay:           bot,
		connectionExchange: connectionExchange,
		serverExchange:     botExchange,
	}
}
