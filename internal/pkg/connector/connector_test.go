package connector

import (
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/relay/connection"
	"github.com/raf924/bot/pkg/users"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
)

type dummyServer struct {
}

func (d *dummyServer) Send(message proto.Message) error {
	panic("implement me")
}

func (d *dummyServer) Recv() (*messages.BotPacket, error) {
	panic("implement me")
}

func (d *dummyServer) Start(botUser *messages.User, users []*messages.User, trigger string) error {
	return nil
}

func (d *dummyServer) Commands() []*messages.Command {
	return []*messages.Command{
		{Name: "echo",
			Aliases: nil,
			Usage:   ""},
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "",
		},
	}
}

type dummyConnection struct {
	users *users.UserList
}

func (d *dummyConnection) Recv() (*messages.MessagePacket, error) {
	panic("implement me")
}

func (d *dummyConnection) Send(message connection.Message) error {
	panic("implement me")
}

func (d *dummyConnection) Start() error {
	return nil
}

func (d *dummyConnection) CommandTrigger() string {
	return "!"
}

func (d *dummyConnection) GetUsers() *users.UserList {
	return d.users.Copy()
}

func (d *dummyConnection) OnUserJoin(f func(user *messages.User, timestamp int64)) {

}

func (d *dummyConnection) OnUserLeft(f func(user *messages.User, timestamp int64)) {

}

func (d *dummyConnection) Connect(nick string) error {
	return nil
}

func TestConnector(t *testing.T) {
	botEx, _ := Cr2BExchange()
	cnEx, _ := Cr2CnExchange()
	bcrEx, _ := B2CrExchange()
	cncrEx, _ := Cn2CrExchange()
	cr := NewConnector(connector.Config{Name: "raf924", Bot: nil, Connection: nil}, &dummyConnection{
		users: users.NewUserList(&messages.User{
			Nick:  "test",
			Id:    "id",
			Mod:   false,
			Admin: false,
		}),
	}, &dummyServer{}, cnEx, botEx)
	err := cr.Start()
	if err != nil {
		t.Errorf("unexpected error = %v", err)
		return
	}
	sentPacket := &messages.MessagePacket{
		Timestamp: timestamppb.Now(),
		Message:   "Hello",
		User: &messages.User{
			Nick:  "test",
			Id:    "id",
			Mod:   false,
			Admin: false,
		},
		Private: false,
	}
	err = cncrEx.Produce(sentPacket)
	if err != nil {
		t.Errorf("unexpected error = %v", err)
	}
	m, err := bcrEx.Consume()
	if err != nil {
		t.Errorf("unexpected errror := %v", err)
	}
	switch m.(type) {
	case *messages.MessagePacket:
		mp := m.(*messages.MessagePacket)
		if mp.GetMessage() != sentPacket.GetMessage() {
			t.Errorf("expected message %v got %v", sentPacket.GetMessage(), mp.GetMessage())
		}
	default:
		t.Errorf("expected MessagePacket")
	}

	if err := bcrEx.Produce(&messages.BotPacket{
		Timestamp: sentPacket.GetTimestamp(),
		Message:   sentPacket.GetMessage(),
		Recipient: sentPacket.GetUser(),
		Private:   sentPacket.GetPrivate(),
	}); err != nil {
		t.Errorf("unexpected error = %v", err)
	}
	m, err = cncrEx.Consume()
	if err != nil {
		t.Errorf("unexpected error = %v", err)
	}
	switch m.(type) {
	case connection.ChatMessage:
		cm := m.(connection.ChatMessage)
		if cm.Message != sentPacket.GetMessage() {
			t.Errorf("expected message %v got %v", sentPacket.GetMessage(), cm.Message)
		}
	default:
		t.Errorf("expected chat message")
	}
}
