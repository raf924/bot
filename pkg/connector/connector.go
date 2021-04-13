package connector

import (
	"github.com/raf924/bot/internal/pkg/connector"
	"github.com/raf924/bot/pkg"
	cnf "github.com/raf924/bot/pkg/config/connector"
	connectionRelay "github.com/raf924/bot/pkg/relay/connection"
	"github.com/raf924/bot/pkg/relay/server"
)

func NewConnector(config cnf.Config) pkg.Runnable {
	withConnectionExchange, err := connector.WithConnectionExchange()
	if err != nil {
		panic(err)
	}
	withBotExchange, err := connector.WithBotExchange()
	if err != nil {
		panic(err)
	}
	connectionExchange, err := connector.ConnectionExchange()
	if err != nil {
		panic(err)
	}
	botExchange, err := connector.BotExchange()
	if err != nil {
		panic(err)
	}
	connection := connectionRelay.GetConnectionRelay(config, connectionExchange)
	bot := server.GetRelayServer(config, botExchange)
	return connector.NewConnector(config, connection, bot, withConnectionExchange, withBotExchange)
}
