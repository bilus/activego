package anycable

import (
	context "context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	reflect "reflect"
	"regexp"

	"github.com/iancoleman/strcase"
	"k8s.io/apimachinery/pkg/util/json"
)

type Channel interface {
	HandleAction(message string) error
}

type ChannelFactory func() (Channel, error)

type TestChannel struct {
}

// TODO: ReflectChannel
func (ch *TestChannel) HandleAction(action string) error {
	// TODO: Handle missing method.
	// TODO: Change snake case to camel case.
	methodName := strcase.ToCamel(action)
	method := reflect.ValueOf(ch).MethodByName(methodName)
	if !method.IsValid() {
		return fmt.Errorf("no such method TestChannel#%v", methodName)
	}
	result := method.Call([]reflect.Value{})
	err := result[0].Interface()
	if err == nil {
		return nil
	}
	return err.(error)
}

func (ch *TestChannel) Tick() error {
	return nil
}

func CreateTestChannel() (Channel, error) {
	return &TestChannel{}, nil
}

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

type CommandData map[string]interface{}

func (c *TestConnection) HandleCommand(identifier, command, data string) error {
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
			case "message":
				channel, err := CreateTestChannel()
				if err != nil {
					log.Printf("Error creating channel: %v", err)
					return false
				}
				parsedData := CommandData{}
				err = json.Unmarshal([]byte(data), &parsedData)
				if err != nil {
					log.Printf("Error parsing data %v: %v", data, err)
					return false
				}
				actionI, ok := parsedData["action"]
				if ok {
					action, ok := actionI.(string)
					if !ok {
						log.Printf("Expecting action to be a string, got this instead: %v", actionI)
						return false
					}
					err = channel.HandleAction(action)
					if err != nil {
						log.Printf("Error handling action %q: %v", action, err)
						return false
					}
				}
				return true

			default:
				log.Printf("No such command %q", command)
				return false
			}
		},
	}
	return testCases.runAll(c)
}
