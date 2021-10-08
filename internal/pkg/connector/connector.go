package connector

import (
	"context"
	"fmt"
	"github.com/raf924/bot/pkg/command"
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/domain"
	"github.com/raf924/bot/pkg/relay/connection"
	"github.com/raf924/bot/pkg/relay/server"
	"log"
	"sort"
	"strings"
	"time"
)

var helpCommand = domain.NewCommand("help", []string{"h"}, "help")

type Connector struct {
	config          connector.Config
	connectionRelay connection.Relay
	relayServer     server.RelayServer
	context         context.Context
	cancelFunc      context.CancelFunc
	users           domain.UserList
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

func (c *Connector) getCommandOr(mP *domain.ChatMessage) domain.ServerMessage {
	log.Printf("Message from %s: %s\n", mP.Sender().Nick(), mP.Message())
	if len(c.config.Trigger) == 0 {
		return mP
	}
	if !strings.HasPrefix(mP.Message(), c.config.Trigger) {
		return mP
	}
	argString := strings.TrimPrefix(mP.Message(), c.config.Trigger)
	args := strings.Split(argString, " ")
	if len(args) == 0 || len(args[0]) == 0 {
		return mP
	}
	possibleCommand := args[0]
	if command.Is(possibleCommand, helpCommand) {
		var names []string
		for _, cmd := range c.relayServer.Commands().All() {
			names = append(names, fmt.Sprintf("%s%s", c.config.Trigger, cmd.Name()))
			for _, alias := range append(cmd.Aliases()) {
				names = append(names, fmt.Sprintf("%s%s (%s)", c.config.Trigger, alias, cmd.Name()))
			}
		}
		sort.Strings(names)
		err := c.sendToConnection(domain.NewClientMessage(strings.Join(names, ", "), mP.Sender(), mP.Private()))
		if err != nil {
			panic(err)
		}
		return nil
	}
	cmd := c.relayServer.Commands().Find(possibleCommand)
	if cmd == nil {
		return mP
	}
	argString = strings.TrimSpace(strings.TrimPrefix(argString, possibleCommand))
	return domain.NewCommandMessage(cmd.Name(), args[1:], argString, mP.Sender(), mP.Private(), mP.Timestamp())
}

func (c *Connector) Start() error {
	c.context, c.cancelFunc = context.WithCancel(context.Background())
	c.connectionRelay.OnUserJoin(func(user *domain.User, timestamp time.Time) {
		err := c.relayServer.Send(domain.NewUserEvent(user, domain.UserJoined, timestamp))
		if err != nil {
			log.Println(err)
		}
	})
	c.connectionRelay.OnUserLeft(func(user *domain.User, timestamp time.Time) {
		err := c.relayServer.Send(domain.NewUserEvent(user, domain.UserLeft, timestamp))
		if err != nil {
			log.Println(err)
		}
	})
	u, users, err := c.connectionRelay.Connect(c.config.Name)
	if err != nil {
		return err
	}
	c.users = domain.ImmutableUserList(users)
	if u == nil {
		u = c.users.Find(c.config.Name)
	}
	if u == nil {
		return fmt.Errorf("couldn't find connector among users")
	}
	err = c.relayServer.Start(u, c.users, c.config.Trigger)
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
			err = c.sendToConnection(packet)
			if err != nil {
				c.cancelFunc()
				return
			}
		}
	}()
	return nil
}

func (c *Connector) receiveFromRelayServer() (*domain.ClientMessage, error) {
	return c.relayServer.Recv()
}

func (c *Connector) receiveFromConnection() (*domain.ChatMessage, error) {
	return c.connectionRelay.Recv()
}

func (c *Connector) sendToServer(m domain.ServerMessage) error {
	return c.relayServer.Send(m)
}

func (c *Connector) sendToConnection(m *domain.ClientMessage) error {
	return c.connectionRelay.Send(m)
}

func NewConnector(config connector.Config, connection connection.Relay, bot server.RelayServer) *Connector {
	return &Connector{
		config:          config,
		connectionRelay: connection,
		relayServer:     bot,
	}
}
