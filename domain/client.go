package domain

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type ClientConfig struct {
	writeWait, pongWait, pingPeriod time.Duration
	maxMessageSize                  int64
}

func NewClientConfig() ClientConfig {
	return ClientConfig{
		writeWait:      writeWait,
		pongWait:       pongWait,
		pingPeriod:     pingPeriod,
		maxMessageSize: maxMessageSize,
	}
}

func (cc ClientConfig) SetShortDeadline() ClientConfig {
	cc.writeWait = 100 * time.Millisecond
	cc.pongWait = 600 * time.Millisecond
	cc.pingPeriod = (cc.pongWait * 9) / 10
  return cc
}

type Client struct {
	user       *User
	connection *websocket.Conn
	channel    *Channel
	eventInbox chan []byte
	config     ClientConfig
}

func NewClient(u *User, ws *websocket.Conn, conf ClientConfig) *Client {

	return &Client{
		user:       u,
		connection: ws,
		channel:    u.channelStore,
		eventInbox: make(chan []byte),
		config:     conf,
	}
}

func (c *Client) EventReceiverPump() {
	ticker := time.NewTicker(c.config.pingPeriod)

	defer func() {
		ticker.Stop()
		c.connection.Close()
	}()

	for {
		select {
		case msg, ok := <-c.eventInbox:
			if !ok {
				c.connection.WriteMessage(websocket.CloseMessage, []byte{})
			}
			w, err := c.connection.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Fatal("failed to return a websocket writer:", err)
			}
			w.Write(msg)

		case <-ticker.C:
			c.connection.SetWriteDeadline(time.Now().Add(c.config.writeWait))
			if err := c.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) eventTransmitterPump() {
	defer func() {
		c.channel.unregister <- c
		c.connection.Close()
	}()

	c.connection.SetReadLimit(c.config.maxMessageSize)
	c.connection.SetReadDeadline(time.Now().Add(c.config.pongWait))
	c.connection.SetPongHandler(func(string) error {
		c.connection.SetReadDeadline(time.Now().Add(c.config.pongWait))
		return nil
	})

	for {
		msg := &Message{}

		err := c.connection.ReadJSON(msg)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading from websocket client: %v", err)
			}
			break
		}

		msg.Sender = c.user
		c.channel.broadcast <- msg
	}
}

type ClientStore interface {
  Add(*Client)
}

type InMemoryClientStore []*Client

func NewInMemoryClientStore() InMemoryClientStore {
  return make(InMemoryClientStore, 0)
}

func (imcs InMemoryClientStore) Add(c *Client) {
  imcs = append(imcs, c)
}

type ClientManager struct {
  clientStore ClientStore
  Authentication Authenticator
  clientConf ClientConfig
}

func NewClientManager(ctx context.Context, conf ClientConfig, retentionPeriod time.Duration) *ClientManager {
  return &ClientManager{
    clientStore: NewInMemoryClientStore(),
    clientConf: conf,
    Authentication: NewOTPRetentionMap(ctx, retentionPeriod),
  }
}

func (cm *ClientManager) ServeClient(w http.ResponseWriter, r *http.Request, channel *Channel) {
  user, err := cm.Authentication.Authenticate(r) 
  if err != nil {
    log.Printf("authentication failed: %s", err)
    return
  }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
    log.Printf("failed to upgrade connection: %s", err)
		return
	}

	user.channelStore = channel
	c := NewClient(user, conn, cm.clientConf)
	c.channel.register <- c

	go c.EventReceiverPump()
	go c.eventTransmitterPump()
}

