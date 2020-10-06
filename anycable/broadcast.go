package anycable

import "encoding/json"

type BroadcastAdapter interface {
	BroadcastRaw(payload interface{}) error
}

type Broadcaster struct {
	adapter BroadcastAdapter
}

func NewBroadcaster(adapter BroadcastAdapter) *Broadcaster {
	return &Broadcaster{adapter}
}

type BroadcastData struct {
	Message interface{} `json:"message"`
}

type Broadcast struct {
	Stream string `json:"stream"`
	Data   string `json:"data"`
}

func (b *Broadcaster) Broadcast(stream string, data interface{}) error {
	bs, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	return b.adapter.BroadcastRaw(Broadcast{
		Stream: stream,
		Data:   string(bs),
	})
}

type CommandBroadcast struct {
	Command string      `json:"command"`
	Payload interface{} `json:"payload"`
}

func (b *Broadcaster) BroadcastCommand(command string, payload interface{}) error {
	return b.adapter.BroadcastRaw(CommandBroadcast{Command: command, Payload: payload})
}
