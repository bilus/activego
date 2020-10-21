package chat

import (
	"activego/anycable"
)

func Connected(c anycable.Connection) error {
	return c.IdentifiedBy("user", c.URL().Query().Get("user"))
}

func Subscribed(c anycable.Connection, ch anycable.Channel) error {
	return ch.StreamFrom("chat")
}

func Message(c anycable.Connection, ch anycable.Channel, data anycable.ActionData) error {
	return ch.Broadcast("chat", data["text"])
}
