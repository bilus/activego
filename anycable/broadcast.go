package anycable

type BroadcastAdapter interface {
	BroadcastRaw(payload interface{}) error
}

type Broadcaster struct {
	adapter BroadcastAdapter
}

func NewBroadcaster(adapter BroadcastAdapter) *Broadcaster {
	return &Broadcaster{adapter}
}

type Broadcast struct {
	Stream string      `json:"stream"`
	Data   interface{} `json:"data"`
}

func (b *Broadcaster) Broadcast(stream string, data interface{}) error {
	return b.adapter.BroadcastRaw(Broadcast{Stream: stream, Data: data})
}

type CommandBroadcast struct {
	Command string      `json:"command"`
	Payload interface{} `json:"payload"`
}

func (b *Broadcaster) BroadcastCommand(command string, payload interface{}) error {
	return b.adapter.BroadcastRaw(CommandBroadcast{Command: command, Payload: payload})
}
