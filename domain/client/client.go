package client

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/DanyPops/logues/domain/channel"
	"github.com/DanyPops/logues/domain/message"
	"github.com/DanyPops/logues/domain/user"
	"github.com/gorilla/websocket"
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

type ConnectionConfig struct {
	writeWait, pongWait, pingPeriod time.Duration
	maxMessageSize                  int64
}

func NewConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		writeWait:      writeWait,
		pongWait:       pongWait,
		pingPeriod:     pingPeriod,
		maxMessageSize: maxMessageSize,
	}
}

func (c ConnectionConfig) SetShortDeadline() ConnectionConfig {
	c.writeWait = 100 * time.Millisecond
	c.pongWait = 600 * time.Millisecond
	c.pingPeriod = (c.pongWait * 9) / 10
	return c
}

type Client struct {
	connection           io.ReadWriteCloser
	user                 user.User
	communicationChannel *channel.Channel
	receiverChannel      chan []byte
	receiverTicker       *time.Ticker
	stopChannel          chan struct{}
}

func NewClient(conn io.ReadWriteCloser, u user.User, ch *channel.Channel) *Client {
	return &Client{
		connection:           conn,
		user:                 u,
		communicationChannel: ch,
		receiverChannel:      make(chan []byte),
		receiverTicker:       time.NewTicker(time.Second * 10),
		stopChannel:          make(chan struct{}),
	}
}

func (c *Client) Start() {
	c.communicationChannel.RegisterReceiver <- c
	go c.receiverReader()
	go c.connectionReader()
}

func (c *Client) Stop() {
	c.stopChannel <- struct{}{}
}

func (c *Client) receiverReader() {
	defer func() {
		c.receiverTicker.Stop()
		c.connection.Close()
	}()

	for {
		select {
		case d, ok := <-c.receiverChannel:
			if !ok {
				c.connection.Close()
				return
			}
			if _, err := c.connection.Write(d); err != nil {
				slog.Error("writing to connection:", err)
				c.connection.Close()
				return
			}

		case <-c.receiverTicker.C:
			if _, err := c.connection.Write([]byte{}); err != nil {
				slog.Error("writing to connection:", err)
				c.connection.Close()
				return

			}

		case <-c.stopChannel:
			slog.Debug("Received stop signal")
			return
		}
	}
}

func (c *Client) connectionReader() {
	defer func() {
		c.communicationChannel.UnregisterReceiver <- c
		c.connection.Close()
	}()

	for {
		msg := message.Message{}

		if err := json.NewDecoder(c.connection).Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}

			if !websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				slog.Error("reading from connection:", err)
			}

			break
		}

		if len(msg.Content) == 0 {
			slog.Error("empty message")
			continue
		}

		msg.Sender = c.user

		c.communicationChannel.BroadcastMessage <- msg
	}
}

func (c *Client) Who() string {
	return c.user.Name
}

func (c *Client) Receive() chan<- []byte {
	return c.receiverChannel
}

type ClientStore interface {
	Add(*Client)
}

type InMemoryClientStore []*Client

func NewInMemoryClientStore() InMemoryClientStore {
	return make(InMemoryClientStore, 0)
}

func (s InMemoryClientStore) Add(c *Client) {
	s = append(s, c)
}

type ClientServer struct {
	clientStore ClientStore
}

func NewClientServer() *ClientServer {
	return &ClientServer{
		clientStore: NewInMemoryClientStore(),
	}
}

func (cs *ClientServer) ServeClient(conn io.ReadWriteCloser, user user.User, ch *channel.Channel) {
	c := NewClient(conn, user, ch)
	go c.Start()
}
