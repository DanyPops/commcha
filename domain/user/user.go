package user

import (
	"github.com/DanyPops/logues/domain/conversation"
)

type User struct {
	name          string
	Subscribed    *conversation.Channel
	messageLatest *conversation.Message
}

func NewUser(name string) *User {
  return &User{
    name: name,
  }
}

func (u User) Name() string {
	return u.name
}

func (u *User) MessageSend(m *conversation.Message) {
	u.Subscribed.MessagesPublish(m)
}

func (u *User) MessageReceive(m *conversation.Message) {
	u.messageLatest = m
}

func (u *User) MessageLatest() *conversation.Message {
	return u.messageLatest
}
