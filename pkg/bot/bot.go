package bot

import (
	"context"
	"github.com/raf924/bot/internal/pkg/bot"
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/bot/permissions"
	botConfig "github.com/raf924/bot/pkg/config/bot"
)

type IBot interface {
	context.Context
	Start() error
}

func HandleCommand(command command.Command) {
	bot.Commands = append(bot.Commands, command)
}

func NewBot(config botConfig.Config) IBot {
	return bot.NewBot(config, permissions.GetManager(config.Bot.Users.Permissions), permissions.GetManager(config.Bot.Commands.Permissions))
}
