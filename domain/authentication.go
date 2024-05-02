package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
)

type Authenticator interface {
  RequestToken(http.ResponseWriter, *User) error
	Authenticate(*http.Request) (*User, error)
}

type OTP struct {
	User    *User
	Key     string
	Created time.Time
}

type OTPRetentionMap map[string]OTP

func (rm OTPRetentionMap) Retention(ctx context.Context, retentionPeriod time.Duration) {
	ticker := time.NewTicker(400 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			for _, otp := range rm {
				if otp.Created.Add(retentionPeriod).Before(time.Now()) {
					delete(rm, otp.Key)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func NewOTPRetentionMap(ctx context.Context, retentionPeriod time.Duration) OTPRetentionMap {
	rm := make(OTPRetentionMap)
	go rm.Retention(ctx, retentionPeriod)
	return rm
}

func (rm OTPRetentionMap) NewOTP(u *User) string {
	o := OTP{
		User:    u,
		Key:     "a",
		Created: time.Now(),
	}

	rm[o.Key] = o
  fmt.Print(rm)
	return o.Key
}

func (rm OTPRetentionMap) RequestToken(w http.ResponseWriter, u *User) error {
  otp := rm.NewOTP(u)

  data := struct{
    OTP string `json:"otp"`
  }{otp}

  json.NewEncoder(w).Encode(data)

  return nil
}

func (rm OTPRetentionMap) Authenticate(r *http.Request) (*User, error) {
	otp := r.URL.Query().Get("otp")

	if otp == "" {
		return nil, errors.New("Empty OTP URL variable")
	}

	if _, ok := rm[otp]; !ok {
    return nil, fmt.Errorf("OTP not found in map: %s", otp)
	}
  
  u := rm[otp].User
	delete(rm, otp)

	return u, nil
}
