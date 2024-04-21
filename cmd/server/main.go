package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/DanyPops/logues/domain/conversation"
	"github.com/DanyPops/logues/domain/user"
)

type loguesServer struct {
	channel *conversation.Channel
	http.Handler
}

func NewLoguesServer(c *conversation.Channel) *loguesServer {
	l := new(loguesServer)

	l.channel = c

	m := http.NewServeMux()
	m.HandleFunc("GET /subscribers", l.ChannelSubscribersList)
	m.HandleFunc("POST /subscribers/{id}", l.ChannelSubscribersAdd)
	m.HandleFunc("GET /messages/latest", l.ChannelMessagesLatest)

	l.Handler = m

	return l
}

func (cs *loguesServer) ChannelSubscribersList(w http.ResponseWriter, r *http.Request) {
	users := cs.channel.SubscribersList()

	json.NewEncoder(w).Encode(users)
	w.WriteHeader(http.StatusOK)
}

func (cs *loguesServer) ChannelSubscribersAdd(w http.ResponseWriter, r *http.Request) {
	s := strings.TrimPrefix(r.URL.Path, "/subscribers/")
	u := user.NewUser(s)
	cs.channel.SubscribersAdd(u)
}

func (cs *loguesServer) ChannelMessagesLatest(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(*cs.channel.MessagesLatest())
	w.WriteHeader(http.StatusOK)
}

func main() {
  s := conversation.NewInMemorySubscriberStore()
	l := NewLoguesServer(conversation.NewChannel(s))
	log.Fatal(http.ListenAndServe(":5000", l))
}
