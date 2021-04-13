package pkg

import "context"

type Runnable interface {
	context.Context
	Start() error
}
