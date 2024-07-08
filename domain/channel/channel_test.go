package channel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DanyPops/logues/domain/message"
)

type mockReceiver struct {
	receive chan []byte
	name    string
}

func NewMockReceiver(s string) *mockReceiver {
	return &mockReceiver{
		receive: make(chan []byte),
		name:    s,
	}
}

func (r mockReceiver) Receive() chan<- []byte {
	return r.receive
}

func (r *mockReceiver) Start() {
	for {
		select {
		case <-r.receive:
		}
	}
}

type WaitingRegistrar struct {
	Registrar
	*sync.WaitGroup
}

func NewWaitingRegistrar() *WaitingRegistrar {
	return &WaitingRegistrar{
		Registrar: make(InMemoryRegistrar),
		WaitGroup: new(sync.WaitGroup),
	}
}

func (r *WaitingRegistrar) Register(rcv Receiver) error {
	err := r.Registrar.Register(rcv)
	r.Done()
	return err
}

func (r *WaitingRegistrar) Unregister(rcv Receiver) error {
	err := r.Registrar.Unregister(rcv)
	r.Done()
	return err
}

type WaitingBroadcaster struct {
	Broadcaster
	*sync.WaitGroup
}

func (b *WaitingBroadcaster) Broadcast(msg message.Message) {
	b.Broadcaster.Broadcast(msg)
	b.Done()
}

func NewWaitingBroadcaster(list func() ([]Receiver, error), evict func(Receiver) error) *WaitingBroadcaster {
	return &WaitingBroadcaster{
		Broadcaster: NewDefaultBroadcaster(list, evict),
		WaitGroup:   new(sync.WaitGroup),
	}
}

func RegistrarReceiversAmountEquelsTo(t *testing.T, reg Registrar, nRcv int) {
	rcvList, err := reg.List()
	if err != nil {
		t.Fatalf("failed to list receivers in channel: %s", err)
	}

	if len(rcvList) != nRcv {
		t.Fatalf("receivers amount doesn't equal the desired amount %d, %d - %v", nRcv, len(rcvList), rcvList)
	}
}

func RegistrarCheckReceiver(t *testing.T, reg Registrar, rcv Receiver) {
	exists, err := reg.Check(rcv)

	if err != nil {
		t.Fatalf("failed to check for receiver: %s", err)
	}

	if !exists {
		t.Errorf("receiver missing: %s", rcv)
	}
}

func RegistrarReceiversEquelTo(t *testing.T, reg Registrar, rcvs []Receiver) {
	for _, rcv := range rcvs {
    RegistrarCheckReceiver(t, reg, rcv)
	}
}

func ChannelRegisterReceivers(t *testing.T, chann *Channel, rcvs []Receiver, registrar *WaitingRegistrar) {
	for i, rcv := range rcvs {
		registrar.Add(1)
		name := fmt.Sprint(i)
		r := NewMockReceiver(name)
		go r.Start()
		rcvs[i] = rcv
		chann.RegisterReceiver <- rcv
	}

	registrar.Wait()

	RegistrarReceiversAmountEquelsTo(t, chann, len(rcvs))

	RegistrarReceiversEquelTo(t, chann, rcvs)
}

func ChannelRegisterNewReceivers(t *testing.T, chann *Channel, rcvs []Receiver, registrar *WaitingRegistrar) {
	for i := range len(rcvs) {
		registrar.Add(1)
		rcv := NewMockReceiver(fmt.Sprint(i))
		go rcv.Start()
		rcvs[i] = rcv
		chann.RegisterReceiver <- rcv
	}

	ChannelRegisterReceivers(t, chann, rcvs, registrar)
}

func ChannelUnregisterReceivers(t *testing.T, chann *Channel, rcvs []Receiver, registrar *WaitingRegistrar) {
	amount := len(rcvs)
	registrar.Add(amount)

	for i := range amount {
		chann.UnregisterReceiver <- rcvs[i]
	}

	registrar.Wait()

	RegistrarReceiversAmountEquelsTo(t, chann, 0)
}

func TestChannelRegistration(t *testing.T) {
	t.Run("Test register & unregister", func(t *testing.T) {
		reg := NewWaitingRegistrar()
    evi := NewInMemoryEvictor(reg.Unregister, 10, time.Second * 10)
		bro := NewDefaultBroadcaster(reg.List, evi.Evict)
		chann := NewChannel(reg, bro)

		go chann.Start()
		defer chann.Stop()

		amount := 10
		rcvs := make([]Receiver, amount)

		ChannelRegisterNewReceivers(t, chann, rcvs, reg)

		ChannelUnregisterReceivers(t, chann, rcvs, reg)
	})
}

func TestChannelEvictUnresponsiveReceiver(t *testing.T) {
	t.Run("Testing eviction for unresponsive receivers", func(t *testing.T) {
		reg := NewWaitingRegistrar()
    evi := NewInMemoryEvictor(reg.Unregister, 0, time.Second * 10)
		bro := NewDefaultBroadcaster(reg.List, evi.Evict)
		chann := NewChannel(reg, bro)

		go chann.Start()
		defer chann.Stop()

		msg := message.Message{Content: "Blocked!"}
		r := NewMockReceiver("Blocker")

		reg.Add(1)
		chann.RegisterReceiver <- r
		reg.Wait()

		RegistrarReceiversAmountEquelsTo(t, chann, 1)

		reg.Add(1)
		chann.BroadcastMessage <- msg
		reg.Wait()

		RegistrarReceiversAmountEquelsTo(t, chann, 0)
	})
}

func TestChannelBroadcasting(t *testing.T) {
	t.Run("Test broadcasting", func(t *testing.T) {
		reg := NewWaitingRegistrar()
    evi := NewInMemoryEvictor(reg.Unregister, 10, 10 * time.Second)
		bro := NewWaitingBroadcaster(reg.List, evi.Evict)
		chann := NewChannel(reg, bro)

		go chann.Start()
		defer chann.Stop()

		amount := 10
		rcvs := make([]Receiver, amount)
		ChannelRegisterNewReceivers(t, chann, rcvs, reg)

		msg := message.Message{Content: "howdy partner ;)"}
		var data bytes.Buffer

		if err := json.NewEncoder(&data).Encode(msg); err != nil {
			t.Fatal("encoding error:", err)
		}

		bro.Add(1)
		chann.BroadcastMessage <- msg
		bro.Wait()

		RegistrarReceiversAmountEquelsTo(t, chann, amount)
	})
}
