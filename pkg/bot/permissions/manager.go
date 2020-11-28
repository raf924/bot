package permissions

import (
	"github.com/raf924/bot/internal/pkg/bot"
	botConfig "github.com/raf924/bot/pkg/config/bot"
)

type ManagerBuilder bot.ManagerBuilder

type PermissionManager interface {
	bot.PermissionManager
}

func Manage(format string, builder bot.ManagerBuilder) {
	bot.PermissionFormats[format] = builder
}

func GetManager(config botConfig.PermissionConfig) PermissionManager {
	builder, ok := bot.PermissionFormats[config.Format]
	if !ok {
		return nil
	}
	return builder(config.Location)
}
