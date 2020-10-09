package anycable

import (
	"encoding/json"
	"fmt"
)

type simpleState struct {
	m             map[string]interface{}
	changedFields map[string]struct{}
}

func NewSimpleState(src map[string]string) (*simpleState, error) {
	state := simpleState{
		m:             make(map[string]interface{}),
		changedFields: make(map[string]struct{}),
	}
	for k, js := range src {
		var v interface{}
		if err := json.Unmarshal([]byte(js), &v); err != nil {
			return nil, err
		}
		state.m[k] = v
	}
	return &state, nil
}

func (state simpleState) Encode(dst map[string]string) error {
	for k, v := range state.m {
		bs, err := json.Marshal(v)
		if err != nil {
			return err
		}
		dst[k] = string(bs)
	}
	return nil
}

func (state simpleState) Get(k string) interface{} {
	return state.m[k]
}

func (state *simpleState) Set(k string, v interface{}) {
	state.changedFields[k] = struct{}{}
	state.m[k] = v
}

func (state *simpleState) UpdateString(k string, f func(string) string) error {
	i, ok := state.m[k]
	if !ok {
		return fmt.Errorf("missing value for key: %v", k)
	}
	s, ok := i.(string)
	if !ok {
		return fmt.Errorf("not a string: value at key: %v", k)
	}
	state.Set(k, f(s))
	return nil
}

func (state *simpleState) UpdateFloat64(k string, f func(float64) float64) error {
	i, ok := state.m[k]
	if !ok {
		return fmt.Errorf("missing value for key: %v", k)
	}
	n, ok := i.(float64)
	if !ok {
		return fmt.Errorf("not a float64: value at key: %v", k)
	}
	state.Set(k, f(n))
	return nil
}

// TODO: UpdateMap, UpdateBool, Update

func (state simpleState) Changes() (map[string]string, error) {
	result := make(map[string]string)
	for k := range state.changedFields {
		v := state.m[k]
		bs, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		result[k] = string(bs)
	}
	return result, nil
}
