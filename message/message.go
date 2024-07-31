package message

import (
	"github.com/DanyPops/logues/user"
)

type Message struct {
  Sender  user.User `json:"user"`
	Content string `json:"content"`
}
