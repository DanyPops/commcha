package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/DanyPops/logues/user"
)

type Authenticator struct {
	UserAuthenticator
	TokenAuthenticator
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserAuthenticator interface {
	AuthenticateCredentials(Credentials) (user.User, error)
}

type EchoUserAuth struct {}

func (a EchoUserAuth) AuthenticateCredentials(creds Credentials) (user.User, error) {
  return user.User{Name: creds.Username}, nil
}

type Token struct {
  Key string `json:"key"`

}

type TokenDecoder interface {
	Decode(t *Token) error
}

type TokenAuthenticator interface {
	NewToken(user.User) (Token, error)
	AuthenticateToken(Token) (user.User, error)
	NewDecoder(*http.Request) TokenDecoder
}

type OTP struct {
	Key     string
	Created time.Time
	User    user.User
}

type OTPDecoder struct {
	r *http.Request
}

func (o OTPDecoder) Decode(t *Token) error {
	otp := o.r.URL.Query().Get("otp")
	if otp == "" {
		return errors.New("Empty OTP URL variable")
	}

	t.Key = otp
	return nil
}

type OTPRetentionMap struct {
	lock   *sync.RWMutex
	otpMap map[string]OTP
}

func NewOTPRetentionMap(retentionPeriod time.Duration) OTPRetentionMap {
	rm := OTPRetentionMap{
		lock:   new(sync.RWMutex),
		otpMap: make(map[string]OTP),
	}

	go rm.Retention(retentionPeriod)
	return rm
}

func (rm OTPRetentionMap) Retention(retentionPeriod time.Duration) {
	ticker := time.NewTicker(400 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
      rm.lock.Lock()
			for _, otp := range rm.otpMap {
				if otp.Created.Add(retentionPeriod).Before(time.Now()) {
					delete(rm.otpMap, otp.Key)
				}
			}
      rm.lock.Unlock()
    // TODO - OTP Context
		// case <-ctx.Done():
		// 	return
    // ODOT
		}
	}
}

func (rm OTPRetentionMap) NewDecoder(r *http.Request) TokenDecoder {
	return OTPDecoder{
		r: r,
	}
}

func (rm OTPRetentionMap) NewToken(u user.User) (Token, error) {
	key := "a"
	o := OTP{
		Key:     key,
		Created: time.Now(),
		User:    u,
	}
  slog.Debug("created new otp","otp", o)
  rm.lock.Lock()
	rm.otpMap[key] = o
  rm.lock.Unlock()
	return Token{key}, nil
}

func (rm OTPRetentionMap) AuthenticateToken(t Token) (user.User, error) {
	key := t.Key
  rm.lock.Lock()
  defer rm.lock.Unlock()
	otp, ok := rm.otpMap[key]
	if !ok {
		return user.User{}, fmt.Errorf("OTP not found in map: %s", otp)
	}
  
  delete(rm.otpMap, key)
	u := otp.User

	return u, nil
}
