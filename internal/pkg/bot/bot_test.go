package bot

import (
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/bot/permissions"
	"github.com/raf924/bot/pkg/config/bot"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
)

var botUser = &messages.User{
	Nick:  "bot",
	Id:    "id",
	Mod:   false,
	Admin: false,
}

var user = &messages.User{
	Nick:  "user",
	Id:    "userId",
	Mod:   false,
	Admin: false,
}

type dummyRelay struct {
}

func (d *dummyRelay) GetUsers() []*messages.User {
	return []*messages.User{botUser, user}
}

func (d *dummyRelay) OnUserJoin(f func(user *messages.User, timestamp int64)) {

}

func (d *dummyRelay) OnUserLeft(f func(user *messages.User, timestamp int64)) {

}

func (d *dummyRelay) Connect(registration *messages.RegistrationPacket) (*messages.User, error) {
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
	execute     func(packet *messages.CommandPacket) ([]*messages.BotPacket, error)
	onChat      func(packet *messages.MessagePacket) ([]*messages.BotPacket, error)
	onUserEvent func(packet *messages.UserPacket) ([]*messages.BotPacket, error)
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

func (t *testCommand) Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	return t.execute(command)
}

func (t *testCommand) OnChat(message *messages.MessagePacket) ([]*messages.BotPacket, error) {
	return t.onChat(message)
}

func (t *testCommand) OnUserEvent(packet *messages.UserPacket) ([]*messages.BotPacket, error) {
	return t.onUserEvent(packet)
}

func (t *testCommand) IgnoreSelf() bool {
	return t.ignoreSelf
}

var commandReply = &messages.BotPacket{
	Timestamp: timestamppb.Now(),
	Message:   "command",
	Recipient: user,
	Private:   false,
}

var messageReply = &messages.BotPacket{
	Timestamp: timestamppb.Now(),
	Message:   "message",
	Recipient: user,
	Private:   false,
}

var userEventReply = &messages.BotPacket{
	Timestamp: timestamppb.Now(),
	Message:   "userEvent",
	Recipient: user,
	Private:   false,
}

func testReply(t testing.TB, p proto.Message, expectedReply *messages.BotPacket) {
	err := WithBotExchange.Produce(p)
	if err != nil {
		t.Fatalf("unexpected error = %v", err)
	}
	bp, err := WithBotExchange.Consume()
	if err != nil {
		t.Fatalf("unexpected error = %v", err)
	}

	var gotReply = bp

	switch gotReply.(type) {
	case *messages.BotPacket:
		if gotReply.(*messages.BotPacket).String() != expectedReply.String() {
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
	}, &dummyPermissionManager{}, &dummyPermissionManager{}, &dummyRelay{}, WithRelayExchange, &testCommand{
		init: func(executor command.Executor) error {
			return nil
		},
		execute: func(packet *messages.CommandPacket) ([]*messages.BotPacket, error) {
			return []*messages.BotPacket{
				commandReply,
			}, nil
		},
		onChat: func(packet *messages.MessagePacket) ([]*messages.BotPacket, error) {
			return []*messages.BotPacket{
				messageReply,
			}, nil
		},
		onUserEvent: func(packet *messages.UserPacket) ([]*messages.BotPacket, error) {
			return []*messages.BotPacket{
				userEventReply,
			}, nil
		},
		ignoreSelf: true,
	})
	err := b.Start()
	if err != nil {
		t.Fatalf("unexpected error = %v", err)
	}
	testReply(t, &messages.MessagePacket{
		Timestamp: timestamppb.Now(),
		Message:   "test",
		User:      user,
		Private:   false,
	}, messageReply)
	testReply(t, &messages.CommandPacket{
		Timestamp: timestamppb.Now(),
		Command:   "test",
		Args:      nil,
		User:      user,
		Private:   false,
		ArgString: "",
	}, commandReply)
	testReply(t, &messages.UserPacket{
		Timestamp: timestamppb.Now(),
		User:      user,
		Event:     0,
	}, userEventReply)
}
