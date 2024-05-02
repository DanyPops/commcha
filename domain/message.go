package domain

import "fmt"

type Message struct {
  Sender  *User `json:"user"`
	Content string `json:"content"`
}

func NewMessage(sender *User, content string) *Message {
	return &Message{
		sender,
		content,
	}
}

func (m Message) String() string {
	return fmt.Sprintf("%s: %s", m.Sender.Name, m.Content)
}
