package conversation

type Channel struct {
	messagesLatest   *Message
	subscribersStore SubscribersStore
}

func NewChannel(s SubscribersStore) *Channel {
	return &Channel{
    subscribersStore: s,
  }
}

func (c *Channel) SubscribersList() []Subscriber {
	return c.subscribersStore.List()
}

func (c *Channel) SubscribersAdd(s Subscriber) {
	c.subscribersStore.Add(s)
}

func (c *Channel) MessagesLatest() *Message {
	return c.messagesLatest
}

func (c *Channel) MessagesPublish(m *Message) {
	for _, sub := range c.subscribersStore.List() {
		sub.MessageReceive(m)
	}
}

type Subscriber interface {
	MessageReceive(*Message)
	MessageLatest() *Message
}

type SubscribersStore interface {
	List() []Subscriber
	Add(Subscriber)
}

func NewInMemorySubscriberStore() *InMemorySubscribersStore {
	return &InMemorySubscribersStore{}
}

type InMemorySubscribersStore []Subscriber

func (s InMemorySubscribersStore) List() []Subscriber {
	return s
}

func (s *InMemorySubscribersStore) Add(sub Subscriber) {
	*s = append(*s, sub)
}
