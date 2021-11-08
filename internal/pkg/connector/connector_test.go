package connector

import (
	"github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/bot/pkg/domain"
	"github.com/raf924/bot/pkg/rpc"
	"testing"
	"time"
)

var _ rpc.ConnectorRelay = (*dummyServer)(nil)

type dummyServer struct {
}

func (d *dummyServer) Send(message domain.ServerMessage) error {
	panic("implement me")
}

func (d *dummyServer) Recv() (*domain.ClientMessage, error) {
	panic("implement me")
}

func (d *dummyServer) Start(botUser *domain.User, onlineUsers domain.UserList, trigger string) error {
	return nil
}

func (d *dummyServer) Commands() domain.CommandList {
	return domain.NewCommandList(
		domain.NewCommand("echo", nil, ""),
		domain.NewCommand("test", []string{"t"}, ""),
	)
}

var _ rpc.ConnectionRelay = (*dummyConnection)(nil)

type dummyConnection struct {
	users domain.UserList
}

func (d *dummyConnection) Recv() (*domain.ChatMessage, error) {
	panic("implement me")
}

func (d *dummyConnection) Send(message *domain.ClientMessage) error {
	panic("implement me")
}

func (d *dummyConnection) OnUserJoin(f func(user *domain.User, timestamp time.Time)) {

}

func (d *dummyConnection) OnUserLeft(f func(user *domain.User, timestamp time.Time)) {

}

func (d *dummyConnection) Connect(nick string) (*domain.User, domain.UserList, error) {
	return domain.NewUser(nick, "", domain.RegularUser), d.users, nil
}

func TestConnector(t *testing.T) {
	//TODO: rewrite
	cr := NewConnector(connector.Config{Name: "test", Bot: nil, Connection: nil}, &dummyConnection{
		users: domain.NewUserList(domain.NewUser("test", "id", domain.RegularUser), domain.NewUser("nick2", "id2", domain.RegularUser)),
	}, &dummyServer{})
	err := cr.Start()
	if err != nil {
		t.Errorf("unexpected error = %v", err)
		return
	}
	sentPacket := domain.NewChatMessage("Hello", domain.NewUser("test", "id", domain.RegularUser), nil, false, false, time.Now(), true)
	go func() {
		err = cr.sendToDispatchers(sentPacket)
		if err != nil {
			t.Errorf("unexpected error = %v", err)
		}
		m, err := cr.receiveFromConnection()
		if err != nil {
			t.Errorf("unexpected errror := %v", err)
		}
		if m.Message() != sentPacket.Message() {
			t.Errorf("expected message %v got %v", sentPacket.Message(), m.Message())
		}

		packetFromBot := domain.NewClientMessage(sentPacket.Message(), sentPacket.Sender(), sentPacket.Private())
		_ = packetFromBot
		var cm rpc.ChatMessage
		if cm.Message != sentPacket.Message() {
			t.Errorf("expected message %v got %v", sentPacket.Message(), cm.Message)
		}
	}()
	<-cr.Done()
}
