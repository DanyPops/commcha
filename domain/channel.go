package domain

import (
	"bytes"
	"encoding/json"
	"log"
)

type Channel struct {
	stop         chan bool
	clients      map[*Client]bool
	broadcast    chan *Message
	register     chan *Client
	unregister   chan *Client
	messageStore []*Message
}

func NewChannel() *Channel {
	return &Channel{
		stop:       make(chan bool),
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (c *Channel) Stop() {
	c.stop <- true
}

func (c *Channel) Start() {
	for {
		select {
		case client := <-c.register:
			c.clients[client] = true
			log.Print("client registered: ", client.user.Name)
		case client := <-c.unregister:
			if _, ok := c.clients[client]; ok {
				log.Print("client unregistered: ", client.user.Name)
				delete(c.clients, client)
			}
		case msg := <-c.broadcast:
			for client := range c.clients {
				select {
				case client.eventInbox <- c.encode(msg):
				default:
					close(client.eventInbox)
					delete(c.clients, client)
				}
			}
		case <-c.stop:
			close(c.broadcast)
			close(c.register)
			close(c.unregister)
			close(c.stop)
			log.Print("channel closed")
			return
		}
	}
}

func (c *Channel) encode(m *Message) []byte {
	var encodedMessage bytes.Buffer
	err := json.NewEncoder(&encodedMessage).Encode(m)
	if err != nil {
		log.Fatal("encoding error:", err)
	}
	return encodedMessage.Bytes()
}
