package connector

import (
	"context"
	"fmt"
	"github.com/raf924/bot/pkg"
	"github.com/raf924/bot/pkg/command"
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/domain"
	"github.com/raf924/bot/pkg/rpc"
	"github.com/segmentio/ksuid"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

var helpCommand = domain.NewCommand("help", []string{"h"}, "help")

type Connector struct {
	config          connector.Config
	dispatchers     sync.Map
	connectionRelay rpc.ConnectionRelay
	relayServer     rpc.ConnectorRelay
	context         context.Context
	cancelFunc      context.CancelFunc
	users           domain.UserList
}

var _ pkg.Runnable = (*Connector)(nil)

func (c *Connector) Done() <-chan struct{} {
	return c.context.Done()
}

func (c *Connector) Err() error {
	return c.context.Err()
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
	commands := domain.NewCommandList()
	c.dispatchers.Range(func(key, value interface{}) bool {
		commands.Append(value.(rpc.Dispatcher).Commands())
		return true
	})
	if command.Is(possibleCommand, helpCommand) {
		var names []string
		for _, cmd := range commands.All() {
			names = append(names, fmt.Sprintf("%s%s", c.config.Trigger, cmd.Name()))
			for _, alias := range append(cmd.Aliases()) {
				names = append(names, fmt.Sprintf("%s%s (%s)", c.config.Trigger, alias, cmd.Name()))
			}
		}
		sort.Strings(names)
		err := c.sendToConnection(domain.NewClientMessage(strings.Join(names, ", "), mP.Sender(), mP.Private()))
		if err != nil {
			log.Println(err)
			c.cancelFunc()
		}
		return nil
	}
	cmd := commands.Find(possibleCommand)
	if cmd == nil {
		return mP
	}
	argString = strings.TrimSpace(strings.TrimPrefix(argString, possibleCommand))
	return domain.NewCommandMessage(cmd.Name(), args[1:], argString, mP.Sender(), mP.Private(), mP.Timestamp())
}

func (c *Connector) Start(ctx context.Context) error {
	c.context, c.cancelFunc = context.WithCancel(ctx)
	c.connectionRelay.OnUserJoin(func(user *domain.User, timestamp time.Time) {
		err := c.sendToDispatchers(domain.NewUserEvent(user, domain.UserJoined, timestamp))
		if err != nil {
			log.Println(err)
		}
	})
	c.connectionRelay.OnUserLeft(func(user *domain.User, timestamp time.Time) {
		err := c.sendToDispatchers(domain.NewUserEvent(user, domain.UserLeft, timestamp))
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
	err = c.relayServer.Start(ctx, u, c.users, c.config.Trigger)
	if err != nil {
		return err
	}
	go func() {
		for c.Err() == nil {
			c.dispatchers.Range(func(key, value interface{}) bool {
				select {
				case <-value.(rpc.Dispatcher).Done():
					c.dispatchers.Delete(key)
					return false
				default:
				}
				return true
			})
		}
	}()
	go func() {
		for c.Err() == nil {
			dispatcher, err := c.relayServer.Accept()
			if err != nil {
				c.cancelFunc()
				return
			}
			newUUID, err := ksuid.NewRandom()
			if err != nil {
				c.cancelFunc()
				return
			}
			c.dispatchers.Store(newUUID.String(), dispatcher)
		}
	}()
	go func() {
		for c.Err() == nil {
			mP, err := c.receiveFromConnection()
			if err != nil {
				c.cancelFunc()
				return
			}
			m := c.getCommandOr(mP)
			if m == nil {
				continue
			}
			err = c.sendToDispatchers(m)
			if err != nil {
				log.Println(err)
			}
		}
	}()
	go func() {
		for c.Err() == nil {
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

func (c *Connector) sendToDispatchers(m domain.ServerMessage) error {
	c.dispatchers.Range(func(key, value interface{}) bool {
		go func() {
			err := value.(rpc.Dispatcher).Dispatch(m)
			if err != nil {
				c.dispatchers.Delete(key)
				log.Println(err)
			}
		}()
		return true
	})
	return nil
}

func (c *Connector) sendToConnection(m *domain.ClientMessage) error {
	return c.connectionRelay.Send(m)
}

func NewConnector(config connector.Config, connection rpc.ConnectionRelay, connectorRelay rpc.ConnectorRelay) *Connector {
	return &Connector{
		config:          config,
		connectionRelay: connection,
		relayServer:     connectorRelay,
	}
}
