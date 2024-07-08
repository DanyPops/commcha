package connection

import (
	"encoding/gob"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var upgrader = NewGorillaUpgrader()

func echo(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		buf := make([]byte, 512)
		_, err := conn.Read(buf)
		if err != nil {
			break
		}

		_, err = conn.Write(buf)
		if err != nil {
			break
		}
	}
}

func TestReadWriteClose(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(echo))
	defer srv.Close()

  conn, err := Dial("ws" + strings.TrimPrefix(srv.URL, "http"))
  if err != nil {
    t.Fatal(err)
  }
	defer conn.Close()

	texts := []string{
		"ohms",
		"dagrewbrw",
		"febrewbrwerxqeb",
	}

	for _, want := range texts {
		if err = gob.NewEncoder(conn).Encode(want); err != nil {
			t.Fatal(err)
		}
    
    var got string
		if err = gob.NewDecoder(conn).Decode(&got); err != nil {
			t.Fatal(err)
		}

    if got != want {
			t.Errorf("%s not equel to %s", got, want)
		}
	}
}
