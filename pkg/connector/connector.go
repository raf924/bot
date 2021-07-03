package connector

import (
	"github.com/raf924/bot/internal/pkg/connector"
	"github.com/raf924/bot/pkg"
	cnf "github.com/raf924/bot/pkg/config/connector"
	connectionRelay "github.com/raf924/bot/pkg/relay/connection"
	"github.com/raf924/bot/pkg/relay/server"
)

func NewConnector(config cnf.Config) pkg.Runnable {
	connection := connectionRelay.GetConnectionRelay(config)
	bot := server.GetRelayServer(config)
	return connector.NewConnector(config, connection, bot)
}
