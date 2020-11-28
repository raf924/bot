package connector

import (
	"github.com/raf924/bot/internal/pkg/connector"
	cnf "github.com/raf924/bot/pkg/config/connector"
)
import "context"

type IConnector interface {
	context.Context
	Start() error
}

func NewConnector(config cnf.Config) IConnector {
	return connector.NewConnector(config)
}
