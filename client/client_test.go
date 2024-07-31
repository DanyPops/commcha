package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DanyPops/logues/channel"
	"github.com/DanyPops/logues/connection"
	"github.com/DanyPops/logues/message"
	"github.com/DanyPops/logues/user"
	"github.com/gorilla/websocket"
)

type LockBuffer struct {
	bytes.Buffer
	*sync.RWMutex
}

func NewLockBuffer() *LockBuffer {
	return &LockBuffer{
		RWMutex: new(sync.RWMutex),
	}
}

func (lb *LockBuffer) Read(b []byte) (int, error) {
	lb.RWMutex.RLock()
	defer lb.RWMutex.RUnlock()
	return lb.Buffer.Read(b)
}

func (lb *LockBuffer) Write(b []byte) (int, error) {
	lb.RWMutex.Lock()
	defer lb.RWMutex.Unlock()
	return lb.Buffer.Write(b)
}

type WaitBuffer struct {
	bytes.Buffer
	*sync.WaitGroup
}

func NewWaitBuffer() *WaitBuffer {
	return &WaitBuffer{
		WaitGroup: new(sync.WaitGroup),
	}
}

func (lb *WaitBuffer) Read(b []byte) (int, error) {
	defer lb.WaitGroup.Done()
	rb, err := lb.Buffer.Read(b)
	return rb, err
}

func (lb *WaitBuffer) Write(b []byte) (int, error) {
	defer lb.WaitGroup.Done()
	rb, err := lb.Buffer.Write(b)
	return rb, err
}

type MockConnection struct {
  ilock sync.RWMutex
	input  io.ReadWriter
  olock sync.RWMutex
	output io.ReadWriter
  clock sync.RWMutex
	close  bool
}

func NewMockConnection(i io.ReadWriter, o io.ReadWriter) *MockConnection {
	return &MockConnection{
		input:  i,
		output: o,
	}
}

func (c *MockConnection) Write(b []byte) (int, error) {
  c.olock.Lock()
  defer c.olock.Unlock()
	if c.close {
		return 0, fmt.Errorf("Connection closed")
	}

	return c.output.Write(b)
}

func (c *MockConnection) Read(b []byte) (int, error) {
  c.olock.RLock()
  defer c.olock.RUnlock()
	if c.close {
		return 0, fmt.Errorf("Connection closed")
	}

	return c.input.Read(b)
}

func (c *MockConnection) Close() error {
  c.olock.Lock()
  defer c.olock.Unlock()
	if c.close {
		return fmt.Errorf("Connection closed")
	}

	c.close = true
	return nil
}

type MockClientServer struct {
	*ClientServer
	upgrader *connection.GorillaUpgrader
	channel  *channel.Channel
}

func NewMockClientServer(reg channel.Registrar, bcast channel.Broadcaster) *MockClientServer {
  ch := channel.NewChannel(reg, bcast)
  go ch.Start()
  return &MockClientServer{
    ClientServer: NewClientServer(),
    upgrader: connection.NewGorillaUpgrader(),
    channel: ch,
  }
}

func (s *MockClientServer) clientServeHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r)
	if err != nil {
		return
	}

	s.ServeClient(conn, user.User{Name: "mock"}, s.channel)
}

func TestSendReceive(t *testing.T) {
	t.Run("Send & receive message on mock connection", func(t *testing.T) {
		lockBuf := NewLockBuffer()
		waitBuf := NewWaitBuffer()
		conn := NewMockConnection(lockBuf, waitBuf)
		reg := make(channel.InMemoryRegistrar)
		evi := channel.NewInMemoryEvictor(reg.Unregister, 10, 10*time.Second)
		bcast := channel.NewDefaultBroadcaster(reg.List, evi.Evict)
		chann := channel.NewChannel(reg, bcast)

		go chann.Start()
    defer chann.Stop()

		u := user.User{Name: "usee"}
    msg := message.Message{Sender: u, Content: "hello"}
		client := NewClient(conn, u, chann)

		go client.Start()
    defer client.Stop()

		waitBuf.Add(1)
		client.Receive() <- []byte{}

		waitBuf.Add(1)
		json.NewEncoder(conn.input).Encode(msg)
		waitBuf.Wait()

		msg.Sender = u
		want, _ := json.Marshal(msg)

		if !bytes.Equal(append(want, '\n'), waitBuf.Bytes()) {
			t.Errorf("Wanted %v\ngot %v", append(want, '\n'), waitBuf.Bytes())
		}
	})
}

func TestClientServer(t *testing.T) {
	t.Run("Gorilla Websocket client server", func(t *testing.T) {
		reg := make(channel.InMemoryRegistrar)
		evi := channel.NewInMemoryEvictor(reg.Unregister, 10, 10*time.Second)
		bro := channel.NewDefaultBroadcaster(reg.List, evi.Evict)
    clientSrv := NewMockClientServer(reg, bro)
		srv := httptest.NewServer(http.HandlerFunc(clientSrv.clientServeHandler))
		defer srv.Close()

		url := "ws" + strings.TrimPrefix(srv.URL, "http")
		wsConn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatal(err)
		}

		conn := connection.NewConnection(wsConn)
		defer conn.Close()
  
    u := user.User{Name: "mock"}
		msg := message.Message{Sender: u, Content: "hello"}
    if err = json.NewEncoder(conn).Encode(msg); err != nil {
      t.Fatal(err)
    }
    
    got := message.Message{}
    if err = json.NewDecoder(conn).Decode(&got); err != nil {
      t.Fatal(err)
    }

    if !reflect.DeepEqual(msg, got) {
      t.Errorf("Got %s, want %s", msg, got)
    }
	})
}
