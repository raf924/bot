package pkg

type Runnable interface {
	Start() error
	Done() <-chan struct{}
	Err() error
}
