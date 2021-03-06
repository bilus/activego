package activego

import (
	context "context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	reflect "reflect"

	"github.com/apex/log"
	"github.com/bilus/activego/anycable"
	"github.com/iancoleman/strcase"
)

type StatelessConnection struct {
	env         *anycable.Env
	request     *http.Request
	identifiers ConnectionIdentifiers

	socket      *Socket
	broadcaster *Broadcaster

	channelFactory ChannelFactory
}

func NewStatelessConnection(c context.Context, env *anycable.Env, socket *Socket, broadcaster *Broadcaster, channelFactory ChannelFactory, identifiers ConnectionIdentifiers) (*StatelessConnection, error) {
	header := http.Header{}
	for key, value := range env.Headers {
		header.Set(key, value)
	}
	u, err := url.Parse(env.Url)
	if err != nil {
		return nil, err
	}
	if identifiers == nil {
		identifiers = make(ConnectionIdentifiers)
	}
	request := http.Request{Header: header, URL: u}
	return &StatelessConnection{
		env:            env,
		request:        &request,
		socket:         socket,
		broadcaster:    broadcaster,
		channelFactory: channelFactory,
		identifiers:    identifiers,
	}, nil
}

// TODO: Handle authorization failure.
func (c *StatelessConnection) HandleOpen() error {
	return c.socket.Write(WelcomeResponseTransmission{
		Type: "welcome",
	})
}

func (c *StatelessConnection) HandleClose(subscriptions []string) error {
	for _, identifier := range subscriptions {
		// TODO: Pass istate properly.
		c.socket.GetIState().Select(identifier)
		channel, err := c.channelFactory(c, identifier, c.socket, c.broadcaster)
		if err != nil {
			log.Errorf("Error creating channel %q: %v", identifier, err)
			continue
		}
		if err := channel.HandleUnsubscribe(); err != nil {
			log.Errorf("Error unsubscribing from channel %q: %v", identifier, err)
		}
	}
	return nil
}

func (c *StatelessConnection) HandleCommand(identifier, command, data string) error {
	channel, err := c.channelFactory(c, identifier, c.socket, c.broadcaster)
	if err != nil {
		return fmt.Errorf("error creating channel: %v", err)
	}

	switch command {
	case "subscribe":
		// TODO: Handle reject (ok, err)
		if err := channel.HandleSubscribe(); err != nil {
			return err
		}
		return c.socket.Write(CommandResponseTransmission{
			Type:       "confirm_subscription",
			Identifier: identifier,
		})
	case "unsubscribe":
		return channel.HandleUnsubscribe()
	case "message":
		parsedData := ActionData{}
		if err = json.Unmarshal([]byte(data), &parsedData); err != nil {
			return fmt.Errorf("error parsing data %v: %v", data, err)
		}
		actionI, ok := parsedData["action"]
		if ok {
			action, ok := actionI.(string)
			if !ok {
				return fmt.Errorf("expecting action to be a string, got: %q", actionI)
			}
			if err = handleAction(channel, action, parsedData); err != nil {
				return fmt.Errorf("error handling action %q: %v", action, err)
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported command %q", command)
	}
}

func handleAction(channel Channel, action string, data ActionData) error {
	// ok, err := callChannelMethod(channel, action, data)
	// if err != nil {
	// 	return err
	// }
	// if !ok {
	return channel.HandleAction(action, data)
	// }
	// return nil
}

func callChannelMethod(channel Channel, action string, data ActionData) (bool, error) {
	methodName := strcase.ToCamel(action)
	method := reflect.ValueOf(channel).MethodByName(methodName)
	if !method.IsValid() {
		return false, nil
	}
	result := method.Call([]reflect.Value{reflect.ValueOf(data)})
	err := result[0].Interface()
	if err == nil {
		return true, nil
	}
	return true, err.(error)
}

func (c *StatelessConnection) Identifiers() ConnectionIdentifiers {
	return c.identifiers
}

func (c *StatelessConnection) IdentifiedBy(key string, value interface{}) error {
	c.identifiers[key] = value
	return nil
}

func (c *StatelessConnection) SaveToConnectionResponse(r *anycable.ConnectionResponse) error {
	return c.socket.SaveToConnectionResponse(r)
}

func (c *StatelessConnection) SaveToCommandResponse(r *anycable.CommandResponse) error {
	return c.socket.SaveToCommandResponse(r)
}

func (c *StatelessConnection) State() State {
	return c.socket.GetCState()
}

func (c *StatelessConnection) URL() *url.URL {
	return c.request.URL
}

func (c *StatelessConnection) Header() http.Header {
	return c.request.Header
}

func (c *StatelessConnection) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

func (c *StatelessConnection) Transmit(data interface{}) error {
	return c.socket.Write(data)
}
