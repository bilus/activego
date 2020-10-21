package anycable

import (
	context "context"
	"fmt"
	"activego/anycable/adapters"
)

type ConnectedHandler func(Connection) error
type DisconnectedHandler func(Connection) error

type ConnectionController struct {
	Connection

	connected    ConnectedHandler
	disconnected DisconnectedHandler

	channels map[string]*ChannelController
}

func (c ConnectionController) HandleOpen() error {
	err := c.Connection.HandleOpen()
	if err != nil {
		return err
	}
	return c.connected(c.Connection)
}

func (c ConnectionController) HandleClose(subscriptions []string) error {
	err := c.Connection.HandleClose(subscriptions)
	if err != nil {
		return err
	}
	return c.disconnected(c.Connection)
}

type SubscribedHandler func(Connection, Channel) error
type UnsubscribedHandler func(Connection, Channel) error
type ActionHandler func(Connection, Channel, ActionData) error

type ChannelController struct {
	Channel

	connection Connection

	subscribed     SubscribedHandler
	unsubscribed   UnsubscribedHandler
	actionHandlers map[string]ActionHandler
}

func (c ChannelController) HandleSubscribe() error {
	return c.subscribed(c.connection, c.Channel)
}

func (c ChannelController) HandleUnsubscribe() error {
	return c.unsubscribed(c.connection, c.Channel)
}

func (c ChannelController) HandleAction(action string, data ActionData) error {
	handler, ok := c.actionHandlers[action]
	if !ok {
		return fmt.Errorf("missing action %q for channel %q", action, c.Channel.Identifier().Channel)
	}
	return handler(c.connection, c, data)
}

type ServerBuilder struct {
	*Server
	connectionController ConnectionController
}

func BuildServer(broadcaster *Broadcaster) *ServerBuilder {
	builder := &ServerBuilder{

		connectionController: ConnectionController{
			Connection:   nil,
			connected:    func(Connection) error { return nil },
			disconnected: func(Connection) error { return nil },
			channels:     make(map[string]*ChannelController),
		},
	}
	builder.Server = NewServer(
		func(
			c context.Context,
			env *Env,
			socket *Socket,
			broadcaster *Broadcaster,
			channelFactory ChannelFactory,
			identifiers ConnectionIdentifiers) (Connection, error) {

			controller := builder.connectionController
			var err error
			controller.Connection, err = NewStatelessConnection(c, env, socket, broadcaster, channelFactory, identifiers)
			if err != nil {
				return nil, err
			}
			return controller, nil
		},
		func(connection Connection,
			identifierJSON string,
			socket *Socket,
			broadcaster *Broadcaster) (Channel, error) {

			identifier := ChannelIdentifier{}
			if err := identifier.Unmarshal([]byte(identifierJSON)); err != nil {
				return nil, err
			}
			controller, ok := builder.connectionController.channels[identifier.Channel]
			if !ok {
				return nil, fmt.Errorf("missing channel %q", identifier.Channel)
			}
			controller.connection = connection
			var err error
			controller.Channel, err = NewStatelessChannel(identifierJSON, socket, broadcaster)
			if err != nil {
				return nil, err
			}
			return controller, nil
		},
		broadcaster)

	return builder
}

func (b *ServerBuilder) Connected(f ConnectedHandler) *ServerBuilder {
	b.connectionController.connected = f
	return b
}

func (b *ServerBuilder) Disconnected(f DisconnectedHandler) *ServerBuilder {
	b.connectionController.disconnected = f
	return b
}

func (b *ServerBuilder) MakeEmbedded() EmbeddedAnycable {
	a := StartEmbedded(b.Server)
	b.Server.SetBroadcaster(NewBroadcaster(adapters.NewEmbeddedBroadcastAdapter(a)))
	return a
}

type ChannelBuilder struct {
	controller *ChannelController
}

func (b *ServerBuilder) Channel(name string) *ChannelBuilder {
	controller := ChannelController{
		connection:     nil,
		Channel:        nil,
		subscribed:     func(Connection, Channel) error { return nil },
		unsubscribed:   func(Connection, Channel) error { return nil },
		actionHandlers: make(map[string]ActionHandler),
	}
	b.connectionController.channels[name] = &controller
	return &ChannelBuilder{&controller}
}

func (b *ChannelBuilder) Subscribed(subscribed SubscribedHandler) *ChannelBuilder {
	b.controller.subscribed = subscribed
	return b
}

func (b *ChannelBuilder) Unsubscribed(unsubscribed UnsubscribedHandler) *ChannelBuilder {
	b.controller.unsubscribed = unsubscribed
	return b
}

func (b *ChannelBuilder) Received(action string, handler ActionHandler) *ChannelBuilder {
	b.controller.actionHandlers[action] = handler
	return b
}
