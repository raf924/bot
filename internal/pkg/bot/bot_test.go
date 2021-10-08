package bot

import (
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/bot/permissions"
	"github.com/raf924/bot/pkg/config/bot"
	"github.com/raf924/bot/pkg/domain"
	"testing"
	"time"
)

var botUser = domain.NewUser("bot", "id", domain.RegularUser)

var user = domain.NewUser("user", "userId", domain.RegularUser)

type dummyRelay struct {
}

func (d *dummyRelay) Send(packet *domain.ClientMessage) error {
	panic("implement me")
}

func (d *dummyRelay) Recv() (domain.ServerMessage, error) {
	panic("implement me")
}

func (d *dummyRelay) GetUsers() []*domain.User {
	return []*domain.User{botUser, user}
}

func (d *dummyRelay) Connect(registration *domain.RegistrationMessage) (*domain.User, error) {
	return botUser, nil
}

func (d *dummyRelay) Done() <-chan struct{} {
	return make(chan struct{})
}

type dummyPermissionManager struct {
}

func (d *dummyPermissionManager) GetPermission(id string) (permissions.Permission, error) {
	return permissions.ADMIN, nil
}

func (d *dummyPermissionManager) SetPermission(id string, permission permissions.Permission) error {
	return nil
}

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

func testReply(t testing.TB, p domain.ServerMessage, expectedReply *domain.ClientMessage) {
	err := WithBotExchange.Produce(p)
	if err != nil {
		t.Fatalf("unexpected error = %v", err)
	}
	bp, err := WithBotExchange.Consume()
	if err != nil {
		t.Fatalf("unexpected error = %v", err)
	}

	var gotReply = bp

	switch gotReply := gotReply.(type) {
	case *domain.ClientMessage:
		if gotReply != expectedReply {
			t.Errorf("expected %v got %v", expectedReply, gotReply)
		}
	default:
		t.Fatalf("unexpected BotPacket")
	}
}

func TestBot(t *testing.T) {
	b := NewBot(bot.Config{
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
	}, &dummyPermissionManager{}, &dummyPermissionManager{}, &dummyRelay{}, &testCommand{
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
		ignoreSelf: true,
	})
	err := b.Start()
	if err != nil {
		t.Fatalf("unexpected error = %v", err)
	}
	testReply(t, domain.NewChatMessage("test", user, nil, false, false, time.Now(), true), messageReply)
	testReply(t, domain.NewCommandMessage("test", nil, "", user, false, time.Now()), commandReply)
	testReply(t, domain.NewUserEvent(user, domain.UserJoined, time.Now()), userEventReply)
}
