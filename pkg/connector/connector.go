package connector

import (
	"github.com/raf924/bot/internal/pkg/connector"
	"github.com/raf924/bot/pkg"
	cnf "github.com/raf924/bot/pkg/config/connector"
	"github.com/raf924/connector-sdk/rpc"
)

func NewConnector(config cnf.Config) pkg.Runnable {
	connection := GetConnectionRelay(config)
	connectorRelay := GetConnectorRelay(config)
	return connector.NewConnector(config, connection, connectorRelay)
}

var _ = NewConnector

func GetConnectorRelay(config cnf.Config) rpc.ConnectorRelay {
	for relayKey, relayConfig := range config.Bot {
		relayBuilder := rpc.GetConnectorRelay(relayKey)
		if relayBuilder != nil {
			return relayBuilder(relayConfig)
		}
	}
	return nil
}

func GetConnectionRelay(config cnf.Config) rpc.ConnectionRelay {
	for relayKey, relayConfig := range config.Bot {
		relayBuilder := rpc.GetConnectionRelay(relayKey)
		if relayBuilder != nil {
			return relayBuilder(relayConfig)
		}
	}
	return nil
}
