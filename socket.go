package activego

import (
	"encoding/json"

	"github.com/bilus/activego/anycable"
)

type State interface {
	Get(k string) interface{}
	Set(k string, v interface{})
	UpdateString(k string, f func(string) string) error
	UpdateFloat64(k string, f func(float64) float64) error
	Changes() (map[string]string, error)
	Select(k string)
}

type Socket struct {
	transmissions      []string
	unsubscribeAll     bool
	newSubscriptions   []string
	newUnsubscriptions []string
	cstate             State
	istate             State
	identifier         *string
}

func NewSocket(env *anycable.Env, disconnect bool) (*Socket, error) {
	var err error
	cstate, err := DecodeSimpleState(env.Cstate)
	if err != nil {
		return nil, err
	}
	var istate State
	if disconnect {
		istate, err = DecodeNestedState(env.Istate)
	} else {
		istate, err = DecodeSimpleState(env.Istate)
	}
	if err != nil {
		return nil, err
	}
	return &Socket{
		cstate: cstate,
		istate: istate,
	}, nil
}

func (s *Socket) Write(t interface{}) error {
	json, err := json.Marshal(t)
	if err != nil {
		return err
	}
	s.transmissions = append(s.transmissions, string(json))
	return nil
}

func (s *Socket) GetCState() State {
	return s.cstate
}

func (s *Socket) GetIState() State {
	return s.istate
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

func (s *Socket) SaveToCommandResponse(r *anycable.CommandResponse) error {
	r.Transmissions = append(r.Transmissions, s.transmissions...)
	r.Streams = append(r.Streams, s.newSubscriptions...)
	r.StoppedStreams = append(r.StoppedStreams, s.newUnsubscriptions...)
	r.StopStreams = r.StopStreams || s.unsubscribeAll
	var err error
	r.Env, err = s.envResponse()
	return err
}

func (s *Socket) SaveToConnectionResponse(r *anycable.ConnectionResponse) error {
	r.Transmissions = append(r.Transmissions, s.transmissions...)
	var err error
	r.Env, err = s.envResponse()
	return err
}

func (s *Socket) envResponse() (*anycable.EnvResponse, error) {
	response := anycable.EnvResponse{
		Cstate: make(map[string]string),
		Istate: make(map[string]string),
	}
	var err error
	response.Cstate, err = s.cstate.Changes()
	if err != nil {
		return nil, err
	}
	response.Istate, err = s.istate.Changes()
	if err != nil {
		return nil, err
	}
	return &response, nil
}
