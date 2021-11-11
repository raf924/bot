package bot

import (
	"github.com/raf924/bot/internal/pkg/bot"
	"github.com/raf924/bot/pkg"
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/bot/permissions"
	botConfig "github.com/raf924/bot/pkg/config/bot"
	"github.com/raf924/bot/pkg/rpc"
	"log"
	"reflect"
)

func HandleCommand(command command.Command) {
	log.Println("Handling", command.Name())
	if reflect.TypeOf(command).Kind() != reflect.Ptr {
		log.Println("command must be a pointer type")
	}
	bot.Commands = append(bot.Commands, command)
}

var _ = HandleCommand

func NewBot(config botConfig.Config) pkg.Runnable {
	var userPermissionManager permissions.PermissionManager
	var commandPermissionManager permissions.PermissionManager
	if config.Users.AllowAll {
		userPermissionManager = permissions.NewNoCheckPermissionManager()
		commandPermissionManager = permissions.NewNoCheckPermissionManager()
	} else {
		userPermissionManager = permissions.GetManager(config.Users.Permissions)
		commandPermissionManager = permissions.GetManager(config.Commands.Permissions)
	}
	return bot.NewBot(
		config,
		userPermissionManager,
		commandPermissionManager,
		rpc.GetDispatcherRelay(config),
		bot.Commands...,
	)
}

var _ = NewBot
