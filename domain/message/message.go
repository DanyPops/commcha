package message

import (
	"github.com/DanyPops/logues/domain/user"
)

type Message struct {
  Sender  user.User `json:"user"`
	Content string `json:"content"`
}
