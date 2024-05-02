package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/DanyPops/logues/domain"
)

type loguesServer struct {
	http.Handler
	channel       *domain.Channel
	clientConfig  domain.ClientConfig
	clientManager domain.ClientManager
}

func (s *loguesServer) webSocket(w http.ResponseWriter, r *http.Request) {
	s.clientManager.ServeClient(w, r, s.channel)
}

func (s *loguesServer) userAuthenticator(w http.ResponseWriter, r *http.Request) {
  // TODO - Check for User in User Store 
  u := domain.NewUser("dani")

  // Handle error
  _ = s.clientManager.Authentication.RequestToken(w, u)

	return
}

func (s *loguesServer) userRegister(w http.ResponseWriter, r *http.Request) {}

func NewServer(ctx context.Context, c *domain.Channel, conf domain.ClientConfig) *loguesServer {
	l := new(loguesServer)
	l.channel = c
  l.clientManager = *domain.NewClientManager(ctx, conf, time.Second * 5)

	m := http.NewServeMux()
	m.HandleFunc("/ws", l.webSocket)
	m.HandleFunc("POST /auth", l.userAuthenticator)
	l.Handler = m

	l.clientConfig = conf

	return l
}

func main() {
  ctx := context.Background()
	c := domain.NewChannel()
	go c.Start()
	conf := domain.NewClientConfig()
	l := NewServer(ctx ,c, conf)
	log.Fatal(http.ListenAndServe(":5000", l))
}
