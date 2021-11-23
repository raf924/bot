package bot

import (
	"github.com/raf924/bot/v2/internal/pkg/bot"
	"github.com/raf924/bot/v2/pkg"
	"github.com/raf924/bot/v2/pkg/bot/permissions"
	botConfig "github.com/raf924/bot/v2/pkg/config/bot"
	"github.com/raf924/connector-sdk/rpc"
)

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
		GetDispatcherRelay(config),
		bot.Commands...,
	)
}

func GetDispatcherRelay(config botConfig.Config) rpc.DispatcherRelay {
	for relayKey, relayConfig := range config.Connector {
		relayBuilder := rpc.GetDispatcherRelay(relayKey)
		if relayBuilder != nil {
			return relayBuilder(relayConfig)
		}
	}
	return nil
}

var _ = NewBot
