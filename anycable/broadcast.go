package anycable

import (
	"encoding/json"

	"github.com/anycable/anycable-go/common"
)

type BroadcastAdapter interface {
	BroadcastRaw(payload interface{}) error
}

type Broadcaster struct {
	adapter BroadcastAdapter
}

func NewBroadcaster(adapter BroadcastAdapter) *Broadcaster {
	return &Broadcaster{adapter}
}

func (b *Broadcaster) Broadcast(stream string, data interface{}) error {
	bs, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	return b.adapter.BroadcastRaw(common.StreamMessage{
		Stream: stream,
		Data:   string(bs),
	})
}

func (b *Broadcaster) BroadcastCommand(command string, payload interface{}) error {
	bs, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return b.adapter.BroadcastRaw(common.RemoteCommandMessage{Command: command, Payload: json.RawMessage(bs)})
}
