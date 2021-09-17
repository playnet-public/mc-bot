package noop

type MessageSender struct{}

func (m MessageSender) SendMessage(_ string) error {
	return nil
}
