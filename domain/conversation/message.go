package conversation

import (
	"fmt"
)

type MessageSender interface {
	MessageSend(*Message)
	Name() string
}

type Message struct {
	Sender MessageSender
	Data   string
}

func NewMessage(sender MessageSender, data string) *Message {
	return &Message{
		sender,
		data,
	}
}

func (m Message) String() string {
	n := m.Sender.Name()
	d := m.Data
	return fmt.Sprintf("%s: %s", n, d)
}
