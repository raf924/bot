package bot

import (
	"fmt"
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/bot/permissions"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"strings"
	"time"
)

type builtinCommand struct {
	command.NoOpCommand
	name    string
	execute func(command *messages.CommandPacket) ([]*messages.BotPacket, error)
}

func (c *builtinCommand) Name() string {
	return c.name
}

func (c *builtinCommand) Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	return c.execute(command)
}

func (b *Bot) verifySender(command *messages.CommandPacket) bool {
	return b.verifyId(command.User.Id)
}

func (b *Bot) verifyOther(command *messages.CommandPacket) bool {
	id := command.Args[0]
	return b.verifyId(id)
}

func (b *Bot) verifyId(id string) bool {
	permission, err := b.userPermissionManager.GetPermission(id)
	if err != nil {
		return false
	}
	return permission != permissions.UNKNOWN
}

func (b *Bot) verify(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	args := command.GetArgs()
	op := args[len(args)-1]
	packet := &messages.BotPacket{}
	switch op {
	case "add":
	case "remove":
	case "list":
	default:
		//TODO: verify self, verify other: create validation struct with User and Valid
	}
	return []*messages.BotPacket{packet}, nil
}

func (b *Bot) ban(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	args := command.GetArgs()
	userToBan := strings.Join(strings.Split(args[0], "@"), "")
	duration, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, err
	}
	banInfo := ban{
		start:    time.Now(),
		duration: time.Duration(duration) * time.Second,
	}
	b.bans[userToBan] = banInfo
	var banEnd string
	if duration < 0 {
		banEnd = "the end of times"
	} else {
		banEnd = banInfo.start.Add(banInfo.duration).UTC().String()
	}
	packet := &messages.BotPacket{
		Timestamp: timestamppb.Now(),
		Message:   fmt.Sprintf("@%s has been banned until %s", userToBan, banEnd),
		Recipient: command.GetUser(),
		Private:   command.GetPrivate(),
	}
	return []*messages.BotPacket{packet}, nil
}
