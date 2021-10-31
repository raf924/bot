package pkg

import "context"

type Runnable interface {
	Start(ctx context.Context) error
	Done() <-chan struct{}
	Err() error
}
