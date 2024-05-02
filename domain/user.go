package domain

type User struct {
	Name         string
	channelStore *Channel
}

func NewUser(name string) *User {
	return &User{
		Name: name,
	}
}

type ChannelStore interface {
	Add(*Channel)
}

type SingleChannelStore Channel

func NewSingleChannelStore() SingleChannelStore {
  var s SingleChannelStore
  return s
}

func (sc SingleChannelStore) Add(c *Channel) {
	sc = SingleChannelStore(*c)
}
