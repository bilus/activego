package anycable

import (
	context "context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type StatelessConnection struct {
	env     *Env
	request *http.Request

	socket      *Socket
	broadcaster *Broadcaster

	channelFactory ChannelFactory
}

func NewStatelessConnection(c context.Context, env *Env, socket *Socket, broadcaster *Broadcaster, channelFactory ChannelFactory) (*StatelessConnection, error) {
	header := http.Header{}
	for key, value := range env.Headers {
		header.Set(key, value)
	}
	u, err := url.Parse(env.Url)
	if err != nil {
		return nil, err
	}
	request := http.Request{Header: header, URL: u}
	return &StatelessConnection{
		env:            env,
		request:        &request,
		socket:         socket,
		broadcaster:    broadcaster,
		channelFactory: channelFactory,
	}, nil
}

func (c *StatelessConnection) HandleOpen() error {
	// TODO: Respond here.
	return nil
}

func (c *StatelessConnection) HandleCommand(identifier, command, data string) error {
	channel, err := c.channelFactory(identifier, c.socket, c.broadcaster)
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
	case "message":
		parsedData := CommandData{}
		if err = json.Unmarshal([]byte(data), &parsedData); err != nil {
			return fmt.Errorf("error parsing data %v: %v", data, err)
		}
		actionI, ok := parsedData["action"]
		if ok {
			action, ok := actionI.(string)
			if !ok {
				return fmt.Errorf("expecting action to be a string, got: %q", actionI)
			}
			if err = channel.HandleAction(action, parsedData); err != nil {
				return fmt.Errorf("error handling action %q: %v", action, err)
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported command %q", command)
	}
}
