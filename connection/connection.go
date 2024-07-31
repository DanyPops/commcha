package connection

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	defaultUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type ConnectionUpgrader interface {
  Upgrade(w http.ResponseWriter, r *http.Request) (io.ReadWriteCloser, error)
}

type GorillaUpgrader struct {
	upgrader websocket.Upgrader
}

func NewGorillaUpgrader() *GorillaUpgrader {
	return &GorillaUpgrader{
		upgrader: defaultUpgrader,
	}
}

func (u GorillaUpgrader) Upgrade(w http.ResponseWriter, r *http.Request) (io.ReadWriteCloser, error) {
	c, err := u.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	conn := NewConnection(c)

	return conn, err
}

type Connection struct {
	conn *websocket.Conn
}

func Dial(url string) (*Connection, error) {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open a ws connection on %s: %v", url, err)
	}

	return NewConnection(ws), nil
}

func NewConnection(conn *websocket.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

func (c *Connection) Read(b []byte) (int, error) {
	msgType, reader, err := c.conn.NextReader()
	if err != nil {
		return 0, err
	}

	if msgType != websocket.TextMessage {
		return 0, fmt.Errorf("wrong message type!")
	}

	n, err := reader.Read(b)
	if err != nil {
		return 0, err
	}

	return n, err
}

func (c *Connection) Write(b []byte) (int, error) {
	writer, err := c.conn.NextWriter(websocket.TextMessage)

	if err != nil {
		return 0, err
	}

	n, err := writer.Write(b)
	if err != nil {
		return 0, err
	}

	if err = writer.Close(); err != nil {
		return 0, err
	}

	return n, err
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

// _, err = upgrader.Upgrade(w, r, nil)
// if err != nil {
// 	log.Printf("failed to upgrade connection: %s", err)
// 	return
// }
