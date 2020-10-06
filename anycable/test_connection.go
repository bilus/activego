package anycable

import (
	context "context"
	"fmt"
	"log"
	reflect "reflect"
	"regexp"

	"github.com/iancoleman/strcase"
	"k8s.io/apimachinery/pkg/util/json"
)

type TestChannel struct {
	socket      *Socket
	broadcaster *Broadcaster
	identifier  string
}

func (ch *TestChannel) HandleSubscribe() error {
	channelIdentifier := ChannelIdentifier{}
	err := json.Unmarshal([]byte(ch.identifier), &channelIdentifier)
	if err != nil {
		return fmt.Errorf("unparsable identifier: %q: %v", ch.identifier, err)
	}
	switch channelIdentifier.Channel {
	case "Anyt::TestChannels::SubscriptionAknowledgementRejectorChannel":
		ch.socket.Write(CommandResponseTransmission{
			Type:       "reject_subscription",
			Identifier: ch.identifier,
		})
	case "Anyt::TestChannels::SubscriptionTransmissionsChannel":
		ch.socket.Write(MessageResponseTransmission{
			Message:    "hello",
			Identifier: ch.identifier,
		})
		ch.socket.Write(MessageResponseTransmission{
			Message:    "world",
			Identifier: ch.identifier,
		})
	case "Anyt::TestChannels::RequestAChannel":
		ch.socket.Subscribe("request_a")
	case "Anyt::TestChannels::RequestBChannel":
		ch.socket.Subscribe("request_b")
	case "Anyt::TestChannels::RequestCChannel":
		ch.socket.Subscribe("request_c")
	}
	return nil
}

func (ch *TestChannel) HandleUnsubscribe() error {
	channelIdentifier := ChannelIdentifier{}
	err := json.Unmarshal([]byte(ch.identifier), &channelIdentifier)
	if err != nil {
		return fmt.Errorf("unparsable identifier: %q: %v", ch.identifier, err)
	}
	switch channelIdentifier.Channel {
	case "Anyt::TestChannels::RequestAChannel":
		ch.broadcaster.Broadcast("request_a", map[string]string{"data": "user left"})
	case "Anyt::TestChannels::RequestBChannel":
		ch.broadcaster.Broadcast("request_b", map[string]string{"data": "user left"})
	case "Anyt::TestChannels::RequestCChannel":
		ch.broadcaster.Broadcast("request_c", map[string]string{"data": "user left"})
		// TODO: "user left#{params[:id].presence}"
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

func (ch *TestChannel) Identifier() string {
	return ch.identifier
}

func (ch *TestChannel) Tick(CommandData) error {
	return ch.socket.Write(MessageResponseTransmission{
		Message:    "tock",
		Identifier: ch.Identifier(),
	})
}

func (ch *TestChannel) Echo(data CommandData) error {
	return ch.socket.Write(MessageResponseTransmission{
		Message: map[string]interface{}{
			"response": data["text"],
		},
		Identifier: ch.Identifier(),
	})
}

func CreateTestChannel(identifier string, socket *Socket, broadcaster *Broadcaster) (Channel, error) {
	return &TestChannel{
		identifier:  identifier,
		socket:      socket,
		broadcaster: broadcaster,
	}, nil
}

type ChannelIdentifier struct {
	Channel string `json:"channel"`
}

type TestConnection struct {
	*StatelessConnection
}

func CreateTestConnection(c context.Context, env *Env, socket *Socket, broadcaster *Broadcaster, channelFactory ChannelFactory) (Connection, error) {
	connection, err := NewStatelessConnection(c, env, socket, broadcaster, channelFactory)
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
