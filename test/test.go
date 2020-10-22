package test

import (
	"fmt"
	"log"
	"regexp"

	"github.com/bilus/activego"
)

func Setup(server *activego.ServerBuilder) {
	server.Connected(Connected)
	channels := []string{
		"Anyt::TestChannels::SubscriptionAknowledgementRejectorChannel",
		"Anyt::TestChannels::SubscriptionTransmissionsChannel",
		"Anyt::TestChannels::SubscriptionPerformMethodsChannel",
		"Anyt::TestChannels::SubscriptionWithParamsChannel",
		"Anyt::TestChannels::SubscriptionAknowledgementChannel",
		"Anyt::TestChannels::RequestChannel",
		"Anyt::TestChannels::RequestAChannel",
		"Anyt::TestChannels::RequestBChannel",
		"Anyt::TestChannels::RequestCChannel",
		"Anyt::TestChannels::SingleStreamChannel",
		"Anyt::TestChannels::BroadcastDataToStreamChannel",
		"Anyt::TestChannels::StreamsWithManyClientsChannel",
		"Anyt::TestChannels::StopStreamsChannel",
		"Anyt::TestChannels::MultipleStreamsChannel",
		"Anyt::TestChannels::ChannelStateChannel",
	}
	for _, name := range channels {
		ch := server.Channel(name)
		ch.Subscribed(Subscribed).Unsubscribed(Unsubscribed)
		ch.Received("tick", Tick)
		ch.Received("unfollow", Unfollow)
		ch.Received("echo", Echo)
	}
}

type TestCases map[string]func() bool

func (testCases TestCases) runAll(c activego.Connection) error {
	testName := c.URL().Query().Get("test")
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

func Connected(c activego.Connection) error {
	// Delegate to actual collection.
	testCases := TestCases{
		"request_url": func() bool {
			ok, err := regexp.MatchString("test=request_url", c.URL().String())
			return ok && err == nil
		},
		"cookies": func() bool {
			username, err := c.Cookie("username")
			if err != nil {
				log.Printf("Error reading username from cookies: %v", err)
				return false
			}
			return username.Value == "john green"

		},
		"headers": func() bool {
			return c.Header().Get("X-Api-Token") == "abc"
		},
		"reasons": func() bool {
			return c.URL().Query().Get("reason") != "unauthorized"
		},
		"uid": func() bool {
			uid := c.URL().Query().Get("uid")
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
		return fmt.Errorf("unauthorized")
	}
	return nil
}

func Subscribed(c activego.Connection, ch activego.Channel) error {
	switch ch.Identifier().Channel {
	case "Anyt::TestChannels::SubscriptionAknowledgementRejectorChannel":
		return ch.Reject()
	case "Anyt::TestChannels::SubscriptionTransmissionsChannel":
		c.Transmit(activego.MessageResponseTransmission{
			Message:    "hello",
			Identifier: ch.IdentifierJSON(), // TODO: Pass Identifier itself.
		})
		c.Transmit(activego.MessageResponseTransmission{
			Message:    "world",
			Identifier: ch.IdentifierJSON(),
		})
	case "Anyt::TestChannels::RequestAChannel":
		ch.StreamFrom("request_a")
	case "Anyt::TestChannels::RequestBChannel":
		ch.StreamFrom("request_b")
	case "Anyt::TestChannels::RequestCChannel":
		ch.StreamFrom("request_c")
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
	case "Anyt::TestChannels::ChannelStateChannel":
		ch.StreamFrom("state_counts")
		// TODO: Wrap it in a DSL.
		state := ch.State()
		state.Set("count", 1)
		state.Set("user", map[string]interface{}{"name": ch.Identifier().Params["name"]})
	}
	return nil
}

func Unsubscribed(c activego.Connection, ch activego.Channel) error {
	switch ch.Identifier().Channel {
	case "Anyt::TestChannels::RequestAChannel":
		ch.Broadcast("request_a", map[string]string{"data": "user left"})
	case "Anyt::TestChannels::RequestBChannel":
		ch.Broadcast("request_b", map[string]string{"data": "user left"})
	case "Anyt::TestChannels::RequestCChannel":
		ch.Broadcast("request_c", map[string]string{"data": fmt.Sprintf("user left%v", ch.Identifier().Params["id"])})
	case "Anyt::TestChannels::ChannelStateChannel":
		if ch.Param("notify_disconnect") == nil {
			return nil
		}
		user := ch.State().Get("user")
		if user == nil {
			return fmt.Errorf("no 'user' in istate")
		}
		name := user.(map[string]interface{})["name"]
		ch.Broadcast("state_counts", map[string]string{"data": fmt.Sprintf("user left: %v", name)})
	}
	return nil
}

func Tick(c activego.Connection, ch activego.Channel, data activego.ActionData) error {
	switch ch.Identifier().Channel {
	case "Anyt::TestChannels::ChannelStateChannel":
		state := ch.State()
		state.UpdateFloat64("count", func(v float64) float64 { return v + 2 })
		user := state.Get("user").(map[string]interface{})
		return c.Transmit(activego.MessageResponseTransmission{
			Message:    map[string]interface{}{"count": state.Get("count"), "name": user["name"]},
			Identifier: ch.IdentifierJSON(),
		})
	default:
		return c.Transmit(activego.MessageResponseTransmission{
			Message:    "tock",
			Identifier: ch.IdentifierJSON(),
		})
	}
}

func Unfollow(c activego.Connection, ch activego.Channel, data activego.ActionData) error {
	return ch.StopStreamFrom(data["name"].(string))
}

func Echo(c activego.Connection, ch activego.Channel, data activego.ActionData) error {
	return c.Transmit(activego.MessageResponseTransmission{
		Message: map[string]interface{}{
			"response": data["text"],
		},
		Identifier: ch.IdentifierJSON(),
	})
}
