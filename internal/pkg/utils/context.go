package utils

import (
	"context"
	"errors"
	messages "github.com/raf924/connector-api/pkg/gen"
)

func DoWithContext(f func() (*messages.BotPacket, error), context context.Context) (*messages.BotPacket, error) {
	ch := make(chan *messages.BotPacket, 1)
	errCh := make(chan error, 1)
	go func() {
		packet, err := f()
		if err != nil {
			errCh <- err
		}
		ch <- packet
	}()
	select {
	case packet := <-ch:
		return packet, nil
	case err := <-errCh:
		return nil, err
	case <-context.Done():
		return nil, errors.New("operation cancelled")
	}
}
