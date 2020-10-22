package activego

type statelessChannel struct {
	socket         *Socket
	broadcaster    *Broadcaster
	identifierJSON string // TODO: Marshal identifier.
	identifier     ChannelIdentifier
}

// TODO: Pass ChannelIdentifier instead of JSON.
func NewStatelessChannel(identifierJSON string, socket *Socket, broadcaster *Broadcaster) (*statelessChannel, error) {
	identifier := ChannelIdentifier{}
	if err := identifier.Unmarshal([]byte(identifierJSON)); err != nil {
		return nil, err
	}
	return &statelessChannel{
		identifierJSON: identifierJSON,
		identifier:     identifier,
		socket:         socket,
		broadcaster:    broadcaster,
	}, nil
}

func (ch *statelessChannel) HandleSubscribe() error {
	return nil
}

func (ch *statelessChannel) HandleUnsubscribe() error {
	return nil
}

func (ch *statelessChannel) HandleAction(action string, data ActionData) error {
	return nil
}

func (ch *statelessChannel) IdentifierJSON() string {
	return ch.identifierJSON
}

func (ch *statelessChannel) Identifier() ChannelIdentifier {
	return ch.identifier
}

func (ch *statelessChannel) StreamFrom(broadcasting string) error {
	ch.socket.Subscribe(broadcasting)
	return nil
}

func (ch *statelessChannel) StopStreamFrom(broadcasting string) error {
	ch.socket.Unsubscribe(broadcasting)
	return nil
}

func (ch *statelessChannel) Broadcast(stream string, data interface{}) error {
	return ch.broadcaster.Broadcast(stream, data)
}

func (ch *statelessChannel) State() State {
	return ch.socket.GetIState()
}

func (ch *statelessChannel) Reject() error {
	return ch.socket.Write(CommandResponseTransmission{
		Type:       "reject_subscription",
		Identifier: ch.IdentifierJSON(),
	})
}

func (ch *statelessChannel) Param(k string) interface{} {
	return ch.identifier.Params[k]
}
