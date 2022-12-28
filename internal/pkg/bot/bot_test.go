package bot

import (
	"context"
	"github.com/raf924/bot/v2/internal/pkg/rpc"
	"github.com/raf924/bot/v2/pkg/bot/permissions"
	"github.com/raf924/bot/v2/pkg/config/bot"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
	"github.com/raf924/queue"
	"testing"
	"time"
)

var botUser = domain.NewUser("bot", "id", domain.RegularUser)

var user = domain.NewUser("user", "userId", domain.RegularUser)

type testCommand struct {
	init        func(executor command.Executor) error
	execute     func(packet *domain.CommandMessage) ([]*domain.ClientMessage, error)
	onChat      func(packet *domain.ChatMessage) ([]*domain.ClientMessage, error)
	onUserEvent func(packet *domain.UserEvent) ([]*domain.ClientMessage, error)
	ignoreSelf  bool
}

func (t *testCommand) Init(bot command.Executor) error {
	return t.init(bot)
}

func (t *testCommand) Name() string {
	return "test"
}

func (t *testCommand) Aliases() []string {
	return []string{"t"}
}

func (t *testCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	return t.execute(command)
}

func (t *testCommand) OnChat(message *domain.ChatMessage) ([]*domain.ClientMessage, error) {
	return t.onChat(message)
}

func (t *testCommand) OnUserEvent(packet *domain.UserEvent) ([]*domain.ClientMessage, error) {
	return t.onUserEvent(packet)
}

func (t *testCommand) IgnoreSelf() bool {
	return t.ignoreSelf
}

var commandReply = domain.NewClientMessage("command", user, false)

var messageReply = domain.NewClientMessage("message", user, false)

var userEventReply = domain.NewClientMessage("userEvent", user, false)

func testReply(t testing.TB, input domain.ServerMessage, inputProducer queue.Producer[domain.ServerMessage], outputConsumer queue.Consumer[*domain.ClientMessage], expectedReply *domain.ClientMessage) {
	err := inputProducer.Produce(input)
	if err != nil {
		t.Fatalf("unexpected error = %v", err)
	}
	bp, err := outputConsumer.Consume(context.Background())
	if err != nil {
		t.Fatalf("unexpected error = %v", err)
	}

	var gotReply = bp
	if gotReply != expectedReply {
		t.Errorf("expected %v got %v", expectedReply, gotReply)
	}
}

func TestBot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	clientMessageQueue := queue.NewQueue[*domain.ClientMessage]()
	serverMessageQueue := queue.NewQueue[domain.ServerMessage]()
	clientMessageConsumer, err := clientMessageQueue.NewConsumer()
	if err != nil {
		t.Fatal(err)
	}
	clientMessageProducer := clientMessageQueue
	serverMessageConsumer, err := serverMessageQueue.NewConsumer()
	if err != nil {
		t.Fatal(err)
	}
	serverMessageProducer := serverMessageQueue
	b := NewBot(
		bot.Config{
			Connector: nil,
			Trigger:   "!",
			ApiKeys:   map[string]string{},
			Users: bot.UserConfig{
				AllowAll: true,
			},
			Commands: bot.CommandConfig{
				Disabled: map[string]bool{
					"ban":    true,
					"verify": true,
				},
				Permissions: bot.PermissionConfig{},
			},
		},
		permissions.NewNoCheckPermissionManager(),
		permissions.NewNoCheckPermissionManager(),
		rpc.NewDefaultDispatcherRelay(ctx, domain.NewUserList(), "!", botUser, clientMessageProducer, serverMessageConsumer),
		command.NewCommandList(&testCommand{
			init: func(executor command.Executor) error {
				return nil
			},
			execute: func(packet *domain.CommandMessage) ([]*domain.ClientMessage, error) {
				return []*domain.ClientMessage{
					commandReply,
				}, nil
			},
			onChat: func(packet *domain.ChatMessage) ([]*domain.ClientMessage, error) {
				return []*domain.ClientMessage{
					messageReply,
				}, nil
			},
			onUserEvent: func(packet *domain.UserEvent) ([]*domain.ClientMessage, error) {
				return []*domain.ClientMessage{
					userEventReply,
				}, nil
			},
			ignoreSelf: false,
		}),
	)
	err = b.Start(context.Background())
	if err != nil {
		t.Fatalf("unexpected error = %v", err)
	}
	testReply(t, domain.NewChatMessage("test", user, nil, false, false, time.Now(), true), serverMessageProducer, clientMessageConsumer, messageReply)
	testReply(t, domain.NewCommandMessage("test", nil, "", user, false, time.Now()), serverMessageProducer, clientMessageConsumer, commandReply)
	testReply(t, domain.NewUserEvent(user, domain.UserJoined, time.Now()), serverMessageProducer, clientMessageConsumer, userEventReply)
}
