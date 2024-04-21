package main

import (
	// "encoding/json"
	"fmt"
	"net/http"
	// "net/http/httptest"
	"reflect"
	"testing"

	"github.com/DanyPops/logues/domain/conversation"
	"github.com/DanyPops/logues/domain/user"
)

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// func newGetChanSubsListRequest() *http.Request {
// 	req, _ := http.NewRequest(http.MethodGet, "/subscribers", nil)
// 	return req
// }

func newPostChanSubsAddRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/subscribers/%s", name), nil)
	return req
}

// func TestAddListSubscribers(t *testing.T) {
// 	t.Run("Add subscribers & returns the subscribers", func(t *testing.T) {
// 		s := conversation.NewInMemorySubscriberStore()
// 		c := conversation.NewChannel(s)
// 		l := NewLoguesServer(c)
//
// 		res := httptest.NewRecorder()
//
//     uDani := user.NewUser("dani")
//     uAviv := user.NewUser("aviv")
// 		req := newPostChanSubsAddRequest(uDani)
// 		l.ChannelSubscribersAdd(res, req)
// 		req = newPostChanSubsAddRequest(uAviv)
// 		l.ChannelSubscribersAdd(res, req)
//
// 		req = newGetChanSubsListRequest()
// 		l.ChannelSubscribersList(res, req)
//
// 		want := []*user.User{uDani, uAviv}
// 		got := []*user.User{}
//
// 		err := json.NewDecoder(res.Body).Decode(&got)
//
// 		if err != nil {
// 			t.Fatalf("Failed to decode response %q to users: %v", res.Body, err)
// 		}
//
// 		if !reflect.DeepEqual(got, want) {
// 			t.Errorf("got %v, want %v", got, want)
// 		}
// 	})
// }

func TestUserSendMsg(t *testing.T) {
	t.Run("Send messages & return the latest on subscriber", func(t *testing.T) {
		s := conversation.NewInMemorySubscriberStore()
		c := conversation.NewChannel(s)

		pub := user.NewUser("dani")
		pub.Subscribed = c

		sub := user.NewUser("daria")
		sub.Subscribed = c

		s.Add(pub)
		s.Add(sub)

		want := conversation.NewMessage(pub, "Welcome!")
		pub.MessageSend(want)

		got := sub.MessageLatest()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}

		got = pub.MessageLatest()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
