package chat

import (
	"fmt"
	"stimulus/anycable"
)

// import (
// 	"stimulus/anycable"
// )

// type Connection struct {
// 	anycable.Connection
// }

// func (c *Connection) HandleOpen() error {
// 	return c.IdentifiedBy("user", c.URL().Query().Get("user"))
// }

// type Channel struct {
// 	anycable.Channel
// }

// // TODO: Support Clone, Before, After.

// func (ch *Channel) HandleSubscribe() error {
// 	return ch.StreamFrom("chat")
// }

// func (ch *Channel) Message(data anycable.ActionData) error {
// 	return ch.Broadcast("chat", data["text"])
// }
func Connected(c anycable.Connection) error {
	return c.IdentifiedBy("user", c.URL().Query().Get("user"))
}

func Subscribed(c anycable.Connection, ch anycable.Channel) error {
	fmt.Println("Connection in chat.Subscribed", c)

	return ch.StreamFrom("chat")
}

func Message(c anycable.Connection, ch anycable.Channel, data anycable.ActionData) error {
	return ch.Broadcast("chat", data["text"])
}
