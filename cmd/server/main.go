package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type loguesServer struct {
  channel *connectionChannel
  http.Handler
}

func NewLoguesServer(cc *connectionChannel) *loguesServer {
  c := new(loguesServer)

  c.channel = cc

  m := http.NewServeMux()
  m.HandleFunc("GET /subscribers" ,c.ChanSubsList)
  m.HandleFunc("POST /subscribers/{id}" ,c.ChanSubsAdd)
  m.HandleFunc("GET /messages/latest" ,c.ChanMsgsLatest)

  c.Handler = m

  return c
}

func (cs *loguesServer) ChanSubsList(w http.ResponseWriter, r *http.Request) {
  users := cs.channel.SubsList()

  json.NewEncoder(w).Encode(users)
  w.WriteHeader(http.StatusOK)
}

func (cs *loguesServer) ChanSubsAdd(w http.ResponseWriter, r *http.Request) {
  s := strings.TrimPrefix(r.URL.Path, "/subscribers/")
  u := &user{Name: s}
  cs.channel.SubsAdd(u)
}

func (cs *loguesServer) ChanMsgsLatest(w http.ResponseWriter, r *http.Request) {
  json.NewEncoder(w).Encode(*cs.channel.MsgsLatest())
  w.WriteHeader(http.StatusOK)
}

type user struct {
  Name string
  chanSubed *connectionChannel
  msgsLatest *message
}

func (u user) name() string {
  return u.Name
}

func (u *user) Send(m *message) {
  u.chanSubed.MsgsPub(m)
}

type SubscribersStore interface {
  List() []*user
  Add(*user)
}

func NewInMemSubStore() *InMemorySubscribersStore {
  return &InMemorySubscribersStore{}
}

type InMemorySubscribersStore []*user

func (s InMemorySubscribersStore) List() []*user {
  return s
}

func (s *InMemorySubscribersStore) Add(sub *user) {
  *s = append(*s, sub)
}

type MsgsSender interface {
  name() string
  Send(*message)
}

type message struct {
  Sender MsgsSender
  Data string
}

func NewMsg(sender MsgsSender, data string) *message {
  return &message{
    sender,
    data,
  }
}

func (m message) String() string {
  n := m.Sender.name()
  d := m.Data
  return fmt.Sprintf("%s: %s", n, d)
}

func NewChanConn() *connectionChannel {
  return &connectionChannel{}
}

type connectionChannel struct {
  msgsLatest *message
  subStore SubscribersStore
}

func (cc *connectionChannel) SubsList() []*user {
  return cc.subStore.List()
}

func (cc *connectionChannel) SubsAdd(s *user) {
  cc.subStore.Add(s)
}

func (cc *connectionChannel) MsgsLatest() *message {
  return cc.msgsLatest
}

func (cc *connectionChannel) MsgsPub(m *message) {
  for _, sub := range cc.subStore.List() {
    sub.msgsLatest = m
  }
}

func main() {
  cc := NewChanConn()
  c := NewLoguesServer(cc)

  log.Fatal(http.ListenAndServe(":5000", c))
}
