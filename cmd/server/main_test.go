package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
  "reflect"
)

func assertResponseBody(t testing.TB, got, want string) {
  t.Helper()
  if got != want {
    t.Errorf("got %q, want %q", got, want)
  }
}

func newGetChanSubsListRequest() *http.Request {
  req, _ := http.NewRequest(http.MethodGet, "/subscribers", nil)
  return req
}

func newPostChanSubsAddRequest(name string) *http.Request {
  req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/subscribers/%s", name), nil)
  return req
}

func TestAddListSubscribers(t *testing.T) {
  t.Run("Add subscribers & returns the subscribers", func(t *testing.T) {
    cc := NewChanConn()
    imss := NewInMemSubStore()
    cc.subStore = imss

    cch := NewLoguesServer(cc)

    res := httptest.NewRecorder()

    req := newPostChanSubsAddRequest("aviv")
    cch.ChanSubsAdd(res, req)
    req = newPostChanSubsAddRequest("dani")
    cch.ChanSubsAdd(res, req)

    req = newGetChanSubsListRequest()
    cch.ChanSubsList(res, req)

    want := []user{{Name: "aviv"},{Name: "dani"}}
    got := []user{}

    err := json.NewDecoder(res.Body).Decode(&got)

    if err != nil {
      t.Fatalf("Failed to decode response %q to users: %v", res.Body, err)
    }

    if !reflect.DeepEqual(got, want) {
      t.Errorf("got %v, want %v", got, want)
    }
  })
}

func TestMsgsLatest(t *testing.T) {
  t.Run("Return the latest message", func(t *testing.T) {
    cc := NewChanConn()
    u := user{Name: "dani"}
    cc.msgsLatest = NewMsg(&u, "hello")
    cch := NewLoguesServer(cc)

    req, _ := http.NewRequest(http.MethodGet, "/messages/latest", nil)
    res := httptest.NewRecorder()

    cch.ChanMsgsLatest(res, req)
  
    want := "dani: hello\n"
    got := message{
      Sender: &user{},
    }
    _ = json.NewDecoder(res.Body).Decode(&got)
    assertResponseBody(t, fmt.Sprintln(got), want)
  })
}

func TestUserSendMsg(t *testing.T) {
  t.Run("Send messages & return the latest on subscriber", func(t *testing.T) {
    cc := NewChanConn()
    pub := user{
      Name: "dani", 
      chanSubed: cc,
    }
    sub := user{
      Name: "daria", 
      chanSubed: cc,
    }
    imss := NewInMemSubStore()
    imss.Add(&pub)
    imss.Add(&sub)
    cc.subStore = imss

    want := NewMsg(&pub, "Welcome!")
    pub.Send(want)

    got := sub.msgsLatest
    if !reflect.DeepEqual(got, want) {
      t.Errorf("got %v, want %v", got, want)
    }

    got = pub.msgsLatest
    if !reflect.DeepEqual(got, want) {
      t.Errorf("got %v, want %v", got, want)
    }
  })
}
