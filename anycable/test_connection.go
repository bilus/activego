package anycable

import (
	context "context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"k8s.io/apimachinery/pkg/util/json"
)

type ChannelIdentifier struct {
	Channel string `json:"channel"`
}

type TestConnection struct {
	// Connection
	env     *Env
	socket  *Socket
	request *http.Request
}

func CreateTestConnection(c context.Context, env *Env, socket *Socket) (Connection, error) {
	header := http.Header{}
	for key, value := range env.Headers {
		header.Set(key, value)
	}
	u, err := url.Parse(env.Url)
	if err != nil {
		return nil, err
	}
	request := http.Request{Header: header, URL: u}
	return &TestConnection{
		env:     env,
		socket:  socket,
		request: &request,
	}, nil
}

type TestCases map[string]func() bool

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

func (c *TestConnection) HandleOpen() error {
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
	return testCases.runAll(c)
}

func (c *TestConnection) HandleCommand(identifier, command string) error {
	testCases := TestCases{
		"*": func() bool {
			channelIdentifier := ChannelIdentifier{}
			err := json.Unmarshal([]byte(identifier), &channelIdentifier)
			if err != nil {
				log.Printf("Unexpected or missing identifier: %v", identifier)
				return false
			}
			switch command {
			case "subscribe":
				if channelIdentifier.Channel == "Anyt::TestChannels::SubscriptionAknowledgementRejectorChannel" {
					c.socket.Write(CommandResponseTransmission{
						Type:       "reject_subscription",
						Identifier: identifier,
					})

					return true
				}
				c.socket.Write(CommandResponseTransmission{
					Type:       "confirm_subscription",
					Identifier: identifier,
				})
				return true
			default:
				log.Printf("No such command %q", command)
				return false
			}
		},
	}
	return testCases.runAll(c)
}
