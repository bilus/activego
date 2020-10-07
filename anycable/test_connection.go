package anycable

import (
	context "context"
	"fmt"
	"log"
	reflect "reflect"
	"regexp"

	"github.com/iancoleman/strcase"
)

type TestChannel struct {
	socket         *Socket
	broadcaster    *Broadcaster
	identifierJSON string // TODO: Marshal identifier.
	identifier     ChannelIdentifier
}

func (ch *TestChannel) HandleSubscribe() error {
	switch ch.Identifier().Channel {
	case "Anyt::TestChannels::SubscriptionAknowledgementRejectorChannel":
		ch.socket.Write(CommandResponseTransmission{
			Type:       "reject_subscription",
			Identifier: ch.identifierJSON,
		})
	case "Anyt::TestChannels::SubscriptionTransmissionsChannel":
		ch.socket.Write(MessageResponseTransmission{
			Message:    "hello",
			Identifier: ch.identifierJSON,
		})
		ch.socket.Write(MessageResponseTransmission{
			Message:    "world",
			Identifier: ch.identifierJSON,
		})
	case "Anyt::TestChannels::RequestAChannel":
		ch.socket.Subscribe("request_a")
	case "Anyt::TestChannels::RequestBChannel":
		ch.socket.Subscribe("request_b")
	case "Anyt::TestChannels::RequestCChannel":
		ch.socket.Subscribe("request_c")
	case "Anyt::TestChannels::SingleStreamChannel":
		ch.StreamFrom("a")
	case "Anyt::TestChannels::BroadcastDataToStreamChannel":
		ch.StreamFrom("a")
	case "Anyt::TestChannels::StreamsWithManyClientsChannel":
		ch.StreamFrom("a")
	case "Anyt::TestChannels::StopStreamsChannel":
		ch.StreamFrom("a")
		ch.StreamFrom("b")
	case "Anyt::TestChannels::MultipleStreamsChannel":
		ch.StreamFrom("a")
		ch.StreamFrom("b")
	}
	return nil
}

func (ch *TestChannel) HandleUnsubscribe() error {
	switch ch.Identifier().Channel {
	case "Anyt::TestChannels::RequestAChannel":
		ch.broadcaster.Broadcast("request_a", map[string]string{"data": "user left"})
	case "Anyt::TestChannels::RequestBChannel":
		ch.broadcaster.Broadcast("request_b", map[string]string{"data": "user left"})
	case "Anyt::TestChannels::RequestCChannel":
		ch.broadcaster.Broadcast("request_c", map[string]string{"data": fmt.Sprintf("user left%v", ch.Identifier().Params["id"])})
	}
	return nil
}

// TODO: ReflectChannel
func (ch *TestChannel) HandleAction(action string, data CommandData) error {
	// TODO: Handle missing method.
	// TODO: Change snake case to camel case.
	methodName := strcase.ToCamel(action)
	method := reflect.ValueOf(ch).MethodByName(methodName)
	if !method.IsValid() {
		return fmt.Errorf("no such method TestChannel#%v", methodName)
	}
	result := method.Call([]reflect.Value{reflect.ValueOf(data)})
	err := result[0].Interface()
	if err == nil {
		return nil
	}
	return err.(error)
}

func (ch *TestChannel) IdentifierJSON() string {
	return ch.identifierJSON
}

func (ch *TestChannel) Identifier() ChannelIdentifier {
	return ch.identifier
}

func (ch *TestChannel) StreamFrom(broadcasting string) error {
	ch.socket.Subscribe(broadcasting)
	return nil
}

func (ch *TestChannel) StopStreamFrom(broadcasting string) error {
	ch.socket.Unsubscribe(broadcasting)
	return nil
}

func (ch *TestChannel) Tick(CommandData) error {
	return ch.socket.Write(MessageResponseTransmission{
		Message:    "tock",
		Identifier: ch.IdentifierJSON(),
	})
}

func (ch *TestChannel) Unfollow(data CommandData) error {
	return ch.StopStreamFrom(data["name"].(string))
}

func (ch *TestChannel) Echo(data CommandData) error {
	return ch.socket.Write(MessageResponseTransmission{
		Message: map[string]interface{}{
			"response": data["text"],
		},
		Identifier: ch.IdentifierJSON(),
	})
}

func CreateTestChannel(identifierJSON string, socket *Socket, broadcaster *Broadcaster) (Channel, error) {
	identifier := ChannelIdentifier{}
	if err := identifier.Unmarshal([]byte(identifierJSON)); err != nil {
		return nil, err
	}
	return &TestChannel{
		identifierJSON: identifierJSON,
		identifier:     identifier,
		socket:         socket,
		broadcaster:    broadcaster,
	}, nil
}

type TestConnection struct {
	*StatelessConnection
}

func CreateTestConnection(c context.Context, env *Env, socket *Socket, broadcaster *Broadcaster, channelFactory ChannelFactory, identifiers ConnectionIdentifiers) (Connection, error) {
	connection, err := NewStatelessConnection(c, env, socket, broadcaster, channelFactory, identifiers)
	if err != nil {
		return nil, err
	}
	return &TestConnection{
		StatelessConnection: connection,
	}, nil
}

type TestCases map[string]func() bool

func (c *TestConnection) HandleOpen() error {
	// Delegate to actual collection.
	testCases := TestCases{
		"request_url": func() bool {
			ok, err := regexp.MatchString("test=request_url", c.request.URL.String())
			return ok && err == nil
		},
		"cookies": func() bool {
			username, err := c.request.Cookie("username")
			if err != nil {
				log.Printf("Error reading username from cookies: %v", err)
				return false
			}
			return username.Value == "john green"

		},
		"headers": func() bool {
			return c.request.Header.Get("X-Api-Token") == "abc"
		},
		"reasons": func() bool {
			return c.request.URL.Query().Get("reason") != "unauthorized"
		},
		"uid": func() bool {
			uid := c.request.URL.Query().Get("uid")
			err := c.IdentifiedBy("uid", uid)
			if err != nil {
				log.Printf("Error calling IdentifiedBy: %v", err)
			}
			return uid != ""
		},
		"*": func() bool {
			return true
		},
	}
	if err := testCases.runAll(c); err != nil {
		c.socket.Write(DisconnectResponseTransmission{
			Type:      "disconnect",
			Reason:    "unauthorized",
			Reconnect: false,
		})

		return err
	}

	return c.StatelessConnection.HandleOpen()
}

func (testCases TestCases) runAll(c *TestConnection) error {
	testName := c.request.URL.Query().Get("test")
	if testName == "" {
		testName = "*" // Special catch all for default behaviour.
	}
	test, ok := testCases[testName]
	if !ok {
		log.Printf("No such test: %q", testName)
		return fmt.Errorf("no such test: %q", testName)
	}
	if success := test(); !success {
		log.Printf("Test %q failed", testName)
		return fmt.Errorf("test %q failed", testName)
	}

	return nil
}
