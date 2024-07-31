package auth

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/DanyPops/logues/user"
)

func TestAuth(t *testing.T) {
	t.Run("Request Token & Authenticate", func(t *testing.T) {
    rp := time.Millisecond * 1
    rm := NewOTPRetentionMap(rp)
    w := httptest.NewRecorder()
    want := user.User{Name: "tester"}

    token1, _ := rm.NewToken(want)
    json.NewEncoder(w).Encode(token1)
    otp := token1.Key
    if otp == "" {
      t.Fatalf("Empty token value")
    }

    url := fmt.Sprintf("http://test.url/?otp=%s", otp)
    r := httptest.NewRequest("POST", url, nil)

    var token2 Token
    if err := rm.NewDecoder(r).Decode(&token2); err != nil {
      t.Fatalf("Failed to decode: %s", err)
    }

    got, err := rm.AuthenticateToken(token2)
    if err != nil {
      t.Fatalf("Failed to authenticate: %s", err)
    }

    if !reflect.DeepEqual(got, want) {
      t.Errorf("Got %v, Want %v", got, want)
    }
	})
}
