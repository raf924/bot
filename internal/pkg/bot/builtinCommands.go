package bot

import (
	"fmt"
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/bot/permissions"
	"github.com/raf924/bot/pkg/domain"
	"strconv"
	"strings"
	"time"
)

type builtinCommand struct {
	command.NoOpCommand
	name    string
	execute func(command *domain.CommandMessage) ([]*domain.ClientMessage, error)
}

func (c *builtinCommand) Name() string {
	return c.name
}

func (c *builtinCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	return c.execute(command)
}

func (b *Bot) verifySender(command *domain.CommandMessage) bool {
	return b.verifyId(command.Sender().Id())
}

func (b *Bot) verifyOther(command *domain.CommandMessage) bool {
	id := command.Args()[0]
	return b.verifyId(id)
}

func (b *Bot) verifyId(id string) bool {
	permission, err := b.userPermissionManager.GetPermission(id)
	if err != nil {
		return false
	}
	return permission != permissions.UNKNOWN
}

func (b *Bot) verify(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	args := command.Args()
	op := args[len(args)-1]
	packet := &domain.ClientMessage{}
	switch op {
	case "add":
	case "remove":
	case "list":
	default:
		//TODO: verify self, verify other: create validation struct with User and Valid
	}
	return []*domain.ClientMessage{packet}, nil
}

func (b *Bot) ban(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	defer b.saveBans()
	args := command.Args()
	if len(args) < 2 {
		return nil, fmt.Errorf("missing args")
	}
	userToBan := strings.TrimLeft(args[0], "@")
	duration, err := time.ParseDuration(args[1])
	if err != nil {
		seconds, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return nil, err
		}
		duration = time.Duration(seconds) * time.Second
	}
	banInfo := ban{
		Start:    time.Now(),
		Duration: duration,
	}
	b.bans[userToBan] = banInfo
	var banEnd string
	if duration < 0 {
		banEnd = "the end of times"
	} else {
		banEnd = banInfo.Start.Add(banInfo.Duration).UTC().String()
	}
	packet := domain.NewClientMessage(fmt.Sprintf("@%s has been banned until %s", userToBan, banEnd), command.Sender(), command.Private())
	return []*domain.ClientMessage{packet}, nil
}
