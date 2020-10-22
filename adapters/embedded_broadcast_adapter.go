package adapters

import (
	"encoding/json"
	"fmt"

	"github.com/anycable/anycable-go/common"
)

type Node interface {
	Broadcast(*common.StreamMessage)
	RemoteDisconnect(*common.RemoteDisconnectMessage)
}

type embeddedBroadcastAdapter struct {
	target Node
}

func NewEmbeddedBroadcastAdapter(node Node) *embeddedBroadcastAdapter {
	return &embeddedBroadcastAdapter{node}
}

func (a *embeddedBroadcastAdapter) BroadcastRaw(payload interface{}) error {
	switch m := payload.(type) {
	case common.RemoteCommandMessage:
		if m.Command == "disconnect" {
			dmsg := common.RemoteDisconnectMessage{}
			if err := json.Unmarshal(m.Payload, &dmsg); err != nil {
				return fmt.Errorf("unable to unmarshal remote command: %w", err)
			}
			a.target.RemoteDisconnect(&dmsg)
		} else {
			return fmt.Errorf("unknown remote command: %s", m.Command)
		}
	case common.StreamMessage:
		a.target.Broadcast(&m)
	default:
		return fmt.Errorf("unrecognized payload type: %t", payload)
	}

	return nil
}
