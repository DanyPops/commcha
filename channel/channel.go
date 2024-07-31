package channel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/DanyPops/logues/message"
)

type Receiver interface {
	Receive() chan<- []byte
}

type Registrar interface {
	Register(Receiver) error
	Unregister(Receiver) error
	Check(Receiver) (bool, error)
	List() ([]Receiver, error)
}

type InMemoryRegistrar map[Receiver]bool

func (r InMemoryRegistrar) Register(rcv Receiver) error {
	exists, err := r.Check(rcv)

	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("receiver already exists: %s", rcv)
	}

	r[rcv] = true
	return nil
}

func (r InMemoryRegistrar) Unregister(rcv Receiver) error {
	exists, err := r.Check(rcv)

	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("receiver doesn't exists: %s", rcv)
	}

	delete(r, rcv)
	return nil
}

func (r InMemoryRegistrar) Check(rcv Receiver) (bool, error) {
	_, ok := r[rcv]
	return ok, nil
}

func (r InMemoryRegistrar) List() ([]Receiver, error) {
	keys := make([]Receiver, len(r))
	i := 0

	for key := range r {
		keys[i] = key
		i++
	}

	return keys, nil
}

type EvictionTracker struct {
	EvictionLevel uint
	TimeRaised    time.Time
}

type Evictor interface {
	Evict(Receiver) error
}

type InMemoryEvictor struct {
	retentioMap            map[Receiver]EvictionTracker
	evict                  func(Receiver) error
	evictionLevelThershold uint
	retentionTime          time.Duration
}

func NewInMemoryEvictor(evict func(Receiver) error, et uint, rt time.Duration) *InMemoryEvictor {
	return &InMemoryEvictor{
		retentioMap:            make(map[Receiver]EvictionTracker),
		evict:                  evict,
		evictionLevelThershold: et,
		retentionTime:          rt,
	}
}

func (e *InMemoryEvictor) Evict(rcv Receiver) error {
	tracker, ok := e.retentioMap[rcv]

	if !ok {
		e.retentioMap[rcv] = EvictionTracker{
			TimeRaised:    time.Now(),
			EvictionLevel: 1,
		}
	}

	if tracker.EvictionLevel < e.evictionLevelThershold {
		tracker.EvictionLevel++
		tracker.TimeRaised = time.Now()

		e.retentioMap[rcv] = tracker

		return nil
	}

	return e.evict(rcv)
}

func (e *InMemoryEvictor) Start() {
	ticker := time.NewTicker(e.retentionTime)
	for {
		select {
		case <-ticker.C:
			for rcv, tracker := range e.retentioMap {
				if tracker.TimeRaised.Add(e.retentionTime).Before(time.Now()) {
					delete(e.retentioMap, rcv)
				}
			}
		}
	}
}

type Broadcaster interface {
	Broadcast(message.Message)
}

type DefaultBroadcaster struct {
	list  func() ([]Receiver, error)
	evict func(Receiver) error
}

func NewDefaultBroadcaster(listReceivers func() ([]Receiver, error), evictReceiver func(Receiver) error) *DefaultBroadcaster {
	return &DefaultBroadcaster{
		list:  listReceivers,
		evict: evictReceiver,
	}
}

func (b *DefaultBroadcaster) Broadcast(msg message.Message) {
	rcvs, err := b.list()
	if err != nil {
		fmt.Println("broadcast receivers list error")
	}

	for _, rcv := range rcvs {
		select {
		case rcv.Receive() <- b.encode(msg):
		default:
			if err := b.evict(rcv); err != nil {
				fmt.Println("evict unresponsive receiver error")
			}
		}
	}
}

func (b DefaultBroadcaster) encode(m message.Message) []byte {
	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(m)
	if err != nil {
		slog.Error("encoding error:", err)
	}

	return data.Bytes()
}

type Channel struct {
	Broadcaster
	Registrar
	BroadcastMessage   chan message.Message
	RegisterReceiver   chan Receiver
	UnregisterReceiver chan Receiver
	stopChannel        chan struct{}
}

func NewChannel(reg Registrar, bcast Broadcaster) *Channel {
	return &Channel{
		Registrar:          reg,
		Broadcaster:        bcast,
		BroadcastMessage:   make(chan message.Message),
		RegisterReceiver:   make(chan Receiver),
		UnregisterReceiver: make(chan Receiver),
		stopChannel:        make(chan struct{}),
	}
}

func NewDefaultChannel() *Channel {
	reg := make(InMemoryRegistrar)
	evi := NewInMemoryEvictor(reg.Unregister, 10, 10*time.Second)
	bcast := NewDefaultBroadcaster(reg.List, evi.Evict)
	return NewChannel(reg, bcast)
}

func (c *Channel) Start() {
	for {
		select {
		case rcv := <-c.RegisterReceiver:
			c.Register(rcv)

		case rcv := <-c.UnregisterReceiver:
			c.Unregister(rcv)

		case msg := <-c.BroadcastMessage:
			c.Broadcast(msg)

		case <-c.stopChannel:
			slog.Debug("Received stop signal")
			return
		}
	}
}

func (c *Channel) Stop() {
	c.stopChannel <- struct{}{}
}
