package anycable

import "encoding/json"

type Socket struct {
	transmissions      []string
	unsubscribeAll     bool
	newSubscriptions   []string
	newUnsubscriptions []string
}

func (s *Socket) Write(t interface{}) error {
	json, err := json.Marshal(t)
	if err != nil {
		return err
	}
	s.transmissions = append(s.transmissions, string(json))
	return nil
}

func (s *Socket) Subscribe(broadcasting string) {
	s.newSubscriptions = append(s.newSubscriptions, broadcasting)
}

func (s *Socket) Unsubscribe(broadcasting string) {
	s.newUnsubscriptions = append(s.newUnsubscriptions, broadcasting)
}

func (s *Socket) UnsubscribeAll() {
	s.unsubscribeAll = true
}

func (s *Socket) SaveToCommandResponse(r *CommandResponse) {
	r.Transmissions = append(r.Transmissions, s.transmissions...)
	r.Streams = append(r.Streams, s.newSubscriptions...)
	r.StoppedStreams = append(r.StoppedStreams, s.newUnsubscriptions...)
	r.StopStreams = r.StopStreams || s.unsubscribeAll
}

func (s *Socket) SaveToConnectionResponse(r *ConnectionResponse) {
	r.Transmissions = append(r.Transmissions, s.transmissions...)
}
