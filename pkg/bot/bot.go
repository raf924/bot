package bot

import (
	"github.com/raf924/bot/internal/pkg/bot"
	"github.com/raf924/bot/pkg"
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/bot/permissions"
	botConfig "github.com/raf924/bot/pkg/config/bot"
	"github.com/raf924/bot/pkg/relay/client"
	"log"
	"reflect"
)

type noCheckPermissionManager struct {
}

func (n *noCheckPermissionManager) GetPermission(id string) (permissions.Permission, error) {
	return permissions.ADMIN, nil
}

func (n *noCheckPermissionManager) SetPermission(id string, permission permissions.Permission) error {
	return nil
}

func HandleCommand(command command.Command) {
	log.Println("Handling", command.Name())
	if reflect.TypeOf(command).Kind() != reflect.Ptr {
		log.Println("command must be a pointer type")
	}
	bot.Commands = append(bot.Commands, command)
}

func NewBot(config botConfig.Config) pkg.Runnable {
	var userPermissionManager permissions.PermissionManager
	var commandPermissionManager permissions.PermissionManager
	if config.Users.AllowAll {
		userPermissionManager = &noCheckPermissionManager{}
		commandPermissionManager = &noCheckPermissionManager{}
	} else {
		userPermissionManager = permissions.GetManager(config.Users.Permissions)
		commandPermissionManager = permissions.GetManager(config.Commands.Permissions)
	}
	return bot.NewBot(
		config,
		userPermissionManager,
		commandPermissionManager,
		client.GetRelayClient(config),
		bot.Commands...,
	)
}
