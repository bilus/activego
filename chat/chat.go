package chat

import (
	"context"
	"stimulus/anycable"
)

type Connection struct {
	anycable.Connection
}

func NewConnection(c context.Context, env *anycable.Env, socket *anycable.Socket, broadcaster *anycable.Broadcaster, channelFactory anycable.ChannelFactory, identifiers anycable.ConnectionIdentifiers) (anycable.Connection, error) {
	conn, err := anycable.NewStatelessConnection(c, env, socket, broadcaster, channelFactory, identifiers)
	return &Connection{
		Connection: conn,
	}, err
}

func (c *Connection) HandleOpen() error {
	err := c.IdentifiedBy("user", c.URL().Query().Get("user"))
	if err != nil {
		return err
	}
	return c.Connection.HandleOpen()
}

func NewChannel(identifierJSON string, socket *anycable.Socket, broadcaster *anycable.Broadcaster) (anycable.Channel, error) {
	ch, err := anycable.NewStatelessChannel(identifierJSON, socket, broadcaster)
	return &Channel{
		Channel:     ch,
		Broadcaster: broadcaster,
	}, err
}

type Channel struct {
	anycable.Channel
	*anycable.Broadcaster
}

func (ch *Channel) HandleSubscribe() error {
	ch.StreamFrom("chat")

	return ch.Channel.HandleSubscribe()
}

func (ch *Channel) Message(data anycable.CommandData) error {
	return ch.Broadcast("chat", data["text"])
}
