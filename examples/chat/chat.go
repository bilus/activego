package chat

import (
	"github.com/bilus/activego"
)

func Connected(c activego.Connection) error {
	return c.IdentifiedBy("user", c.URL().Query().Get("user"))
}

func Subscribed(c activego.Connection, ch activego.Channel) error {
	return ch.StreamFrom("chat")
}

func Message(c activego.Connection, ch activego.Channel, data activego.ActionData) error {
	return ch.Broadcast("chat", data["text"])
}
