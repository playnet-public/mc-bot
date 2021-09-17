package noop

import "context"

type MessageSender struct{}

func (m MessageSender) SendMessage(_ context.Context, _ string) error {
	return nil
}
