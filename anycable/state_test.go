package anycable_test

import (
	"activego/anycable"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestState_SimpleState_FromNil(t *testing.T) {
	require := require.New(t)

	state, err := anycable.DecodeSimpleState(nil)
	require.NoError(err)
	require.Nil(state.Get("foo"))
}

func TestState_SimpleState_FromMap(t *testing.T) {
	require := require.New(t)

	state, err := anycable.DecodeSimpleState(map[string]string{"foo": `"bar"`})
	require.NoError(err)
	require.Equal("bar", state.Get("foo"))
}

func TestState_SimpleState_Set(t *testing.T) {
	require := require.New(t)

	state, err := anycable.DecodeSimpleState(map[string]string{"foo": `"bar"`})
	require.NoError(err)
	state.Set("foo", "qux")
	require.Equal("qux", state.Get("foo"))
}

func TestState_SimpleState_Update_CorrectType(t *testing.T) {
	require := require.New(t)

	state, err := anycable.DecodeSimpleState(map[string]string{"foo": `"bar"`})
	require.NoError(err)
	err = state.UpdateString("foo", func(v string) string { return v + "BAR" })
	require.NoError(err)
	require.Equal("barBAR", state.Get("foo"))
}

func TestState_SimpleState_Update_IncorrectType(t *testing.T) {
	require := require.New(t)

	state, err := anycable.DecodeSimpleState(map[string]string{"foo": `"bar"`})
	require.NoError(err)
	err = state.UpdateFloat64("foo", func(v float64) float64 { return v + 1 })
	require.Error(err)
	require.Equal("bar", state.Get("foo"))
}

func TestState_SimpleState_Changes_NoChanges(t *testing.T) {
	require := require.New(t)

	state, err := anycable.DecodeSimpleState(map[string]string{"foo": `"bar"`})
	require.NoError(err)
	changes, err := state.Changes()
	require.NoError(err)
	require.Empty(changes)
}

func TestState_SimpleState_Changes_Set(t *testing.T) {
	require := require.New(t)

	state, err := anycable.DecodeSimpleState(map[string]string{"foo": `"bar"`, "baz": `"qux"`})
	require.NoError(err)
	state.Set("baz", "XXX")
	changes, err := state.Changes()
	require.NoError(err)
	require.Equal(map[string]string{"baz": `"XXX"`}, changes)
}

func TestState_SimpleState_Changes_Update(t *testing.T) {
	require := require.New(t)

	state, err := anycable.DecodeSimpleState(map[string]string{"foo": `"bar"`, "baz": `"qux"`})
	require.NoError(err)
	err = state.UpdateString("baz", func(string) string { return "XXX" })
	require.NoError(err)
	changes, err := state.Changes()
	require.NoError(err)
	require.Equal(map[string]string{"baz": `"XXX"`}, changes)
}

func TestState_NestedState_SelectIState(t *testing.T) {
	require := require.New(t)

	state, err := anycable.DecodeNestedState(map[string]string{"foo": `{"bar": "baz"}`})
	require.NoError(err)
	state.Select("foo")
	err = state.UpdateString("bar", func(string) string { return "XXX" })
	require.NoError(err)
	require.Equal("XXX", state.Get("bar"))
	changes, err := state.Changes()
	require.NoError(err)
	require.Equal(map[string]string{"foo": `{"bar": "XXX"}`}, changes)
}
