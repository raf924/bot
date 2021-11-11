package connector

import (
	"github.com/raf924/bot/internal/pkg/connector"
	"github.com/raf924/bot/pkg"
	cnf "github.com/raf924/bot/pkg/config/connector"
	connectionRelay "github.com/raf924/bot/pkg/rpc"
)

func NewConnector(config cnf.Config) pkg.Runnable {
	connection := connectionRelay.GetConnectionRelay(config)
	connectorRelay := connectionRelay.GetConnectorRelay(config)
	return connector.NewConnector(config, connection, connectorRelay)
}

var _ = NewConnector
