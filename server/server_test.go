package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/DanyPops/logues/auth"
	"github.com/DanyPops/logues/connection"
	"github.com/DanyPops/logues/message"
	"github.com/DanyPops/logues/user"
)

type loguesClient struct {
	connection io.ReadWriteCloser
	msgLog     []message.Message
	wg         sync.WaitGroup
}

func connect(url string, creds auth.Credentials) (*loguesClient, error) {
	otp, err := getOTP(url, creds)
	if err != nil {
		return nil, err
	}

	wsURL := "ws" + strings.TrimPrefix(url, "http") + "/ws?otp=" + otp
	conn, err := connection.Dial(wsURL)
	if err != nil {
		return nil, err
	}

	c := &loguesClient{
		connection: conn,
		msgLog:     make([]message.Message, 0),
	}

	go c.Start()

	return c, nil
}

func getOTP(url string, creds auth.Credentials) (string, error) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(creds)
	req, err := http.NewRequest("POST", url+"/auth", body)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var otp auth.Token
	json.NewDecoder(resp.Body).Decode(&otp)

	return otp.Key, nil
}

func (c *loguesClient) Start() {
	for {
		// TODO - MSG R
		msg := message.Message{}

		if err := json.NewDecoder(c.connection).Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}

			if !websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				fmt.Printf("reading from connection: %s", err)
			}

			break
		}

		c.msgLog = append(c.msgLog, msg)
		// ODOT
		c.wg.Done()
	}
}

func (c *loguesClient) SendMessage(content string) error {
	return json.NewEncoder(c.connection).Encode(message.Message{Content: content})
}

func (c *loguesClient) LastMessage() message.Message {
	l := len(c.msgLog)
	if l == 0 {
		return message.Message{}
	}
	return c.msgLog[l-1]
}

func TestServer(t *testing.T) {
	srv := httptest.NewServer(New())
	defer srv.Close()

	t.Run("verify static pages", func(t *testing.T) {
		resp, err := http.Get(srv.URL)

		if err != nil {
			t.Fatal(err)
		}

		func(t *testing.T, got, want int) {
			if got != want {
				t.Errorf("got status code %d, want %d", got, want)
			}
		}(t, resp.StatusCode, http.StatusOK)

		func(t *testing.T, got string) {
      want := "text/html; charset=utf-8"
			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		}(t, resp.Header.Get("Content-Type"))
	})
	t.Run("authenticate user & send message", func(t *testing.T) {
		name := "dpop"
		creds := auth.Credentials{
			Username: name,
			Password: "",
		}
		content := "Hello!"
		want := message.Message{
			Sender:  user.User{Name: name},
			Content: content,
		}
		clis := make([]*loguesClient, 1000)
		for i := range len(clis) {
			c, err := connect(srv.URL, creds)
			if err != nil {
				t.Fatal(err)
			}
			clis[i] = c
			c.wg.Add(1)
		}

		clis[0].SendMessage(content)

		for _, c := range clis {
			c.wg.Wait()
			got := c.LastMessage()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %s, want %s", got, want)
			}
		}
	})
}
