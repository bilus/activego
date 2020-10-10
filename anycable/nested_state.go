package anycable

import (
	"encoding/json"
)

type nestedState struct {
	m     map[string]*simpleState
	focus *simpleState // TODO: Brittle, have to call Select before anything else.
}

func DecodeNestedState(src map[string]string) (*nestedState, error) {
	state := nestedState{
		m: make(map[string]*simpleState),
	}
	for k, js := range src {
		// fmt.Println("istate at", k)
		// spew.Dump(js)
		var m map[string]string
		if err := json.Unmarshal([]byte(js), &m); err != nil {
			return nil, err
		}
		// fmt.Println("Unmarshaled")
		// spew.Dump(m)
		var err error
		state.m[k], err = DecodeSimpleState(m)
		if err != nil {
			return nil, err
		}
	}
	return &state, nil
}

func (state nestedState) Get(k string) interface{} {
	return state.focus.Get(k)
}

func (state *nestedState) Set(k string, v interface{}) {
	state.focus.Set(k, v)
}

func (state *nestedState) UpdateString(k string, f func(string) string) error {
	return state.focus.UpdateString(k, f)
}

func (state *nestedState) UpdateFloat64(k string, f func(float64) float64) error {
	return state.focus.UpdateFloat64(k, f)
}

// TODO: UpdateMap, UpdateBool, Update

func (state nestedState) Changes() (map[string]string, error) {
	result := make(map[string]string)
	for k := range state.m {
		s := state.m[k]
		changes, err := s.RawChanges()
		if err != nil {
			return nil, err
		}
		bs, err := json.Marshal(changes)
		if err != nil {
			return nil, err
		}
		result[k] = string(bs)
	}
	return result, nil
}

func (state *nestedState) Select(k string) {
	var ok bool
	state.focus, ok = state.m[k]
	if !ok {
		state.focus = NewSimpleState(nil)
		state.m[k] = state.focus
	}
}
