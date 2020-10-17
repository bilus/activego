package anycable

import (
	context "context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/metrics"
	"github.com/anycable/anycable-go/node"
)

type EmbeddedAnycable struct {
	appNode *node.Node
	metrics *metrics.Metrics
	http.Handler
}

// TODO: Shutdown it in main.go
func (e EmbeddedAnycable) Shutdown() {
	e.appNode.Shutdown()
	e.metrics.Shutdown()
}

func (e EmbeddedAnycable) Broadcast(m *common.StreamMessage) {
	e.appNode.Broadcast(m)
}

func (e EmbeddedAnycable) RemoteDisconnect(m *common.RemoteDisconnectMessage) {
	e.appNode.RemoteDisconnect(m)
}

func StartEmbedded(server *Server) EmbeddedAnycable {
	controller := NewController(server)
	metrics := metrics.NewMetrics(metrics.NewBasePrinter(), 15)
	appNode := node.NewNode(controller, metrics)
	disconnector := node.NewDisconnectQueue(appNode, &node.DisconnectQueueConfig{
		Rate:            100,
		ShutdownTimeout: 5,
	})
	appNode.Start()
	go disconnector.Run() // nolint:errcheck
	appNode.SetDisconnector(disconnector)

	headers := []string{"cookies"} // TODO: Make it configurable.
	wsConfig := node.NewWSConfig() // TODO: Make it configurable.
	return EmbeddedAnycable{
		appNode: appNode,
		metrics: metrics,
		Handler: node.WebsocketHandler(appNode, headers, &wsConfig),
	}
}

type Controller struct {
	server *Server
}

func NewController(server *Server) *Controller {
	return &Controller{server}
}

func (c *Controller) Shutdown() error {
	return nil
}

// TODO: Make sure that everything is thread-safe in all methods below (test!).

func (c *Controller) Authenticate(sid string, env *common.SessionEnv) (*common.ConnectResult, error) {
	r, err := c.server.Connect(newContext(sid), &ConnectionRequest{
		Path:    env.URL,
		Headers: *env.Headers,
		Env:     buildEnv(env),
	})
	if err != nil {
		return nil, err
	}

	reply := common.ConnectResult{Transmissions: r.Transmissions}

	if r.Env != nil {
		reply.CState = r.Env.Cstate
	}

	if r.Status.String() == "SUCCESS" {
		reply.Identifier = r.Identifiers
		return &reply, nil
	}

	return &reply, fmt.Errorf("Application error: %s", r.ErrorMsg)
}

func (c *Controller) Subscribe(sid string, env *common.SessionEnv, id string, channel string) (*common.CommandResult, error) {
	r, err := c.server.Command(newContext(sid), &CommandMessage{
		Command:               "subscribe",
		Env:                   buildChannelEnv(channel, env),
		Identifier:            channel,
		ConnectionIdentifiers: id},
	)

	return c.parseCommandResponse(r, err)
}

func (c *Controller) Unsubscribe(sid string, env *common.SessionEnv, id string, channel string) (*common.CommandResult, error) {
	r, err := c.server.Command(newContext(sid), &CommandMessage{
		Command:               "unsubscribe",
		Env:                   buildChannelEnv(channel, env),
		Identifier:            channel,
		ConnectionIdentifiers: id,
	})
	return c.parseCommandResponse(r, err)
}

func (c *Controller) Perform(sid string, env *common.SessionEnv, id string, channel string, data string) (*common.CommandResult, error) {
	r, err := c.server.Command(newContext(sid), &CommandMessage{
		Command:               "message",
		Env:                   buildChannelEnv(channel, env),
		Identifier:            channel,
		ConnectionIdentifiers: id,
		Data:                  data,
	})

	return c.parseCommandResponse(r, err)
}

func (c *Controller) Disconnect(sid string, env *common.SessionEnv, id string, subscriptions []string) error {
	r, err := c.server.Disconnect(newContext(sid), &DisconnectRequest{
		Identifiers:   id,
		Subscriptions: subscriptions,
		Path:          env.URL,
		Headers:       *env.Headers,
		Env:           buildDisconnectEnv(env),
	})

	if err != nil {
		return err
	}

	if r.Status.String() == "SUCCESS" {
		return nil
	}

	return fmt.Errorf("Application error: %s", r.ErrorMsg)
}

func newContext(sessionID string) context.Context {
	return context.Background()
	// TODO: I don't think we need it but verify it before production-ready.
	// md := metadata.Pairs("sid", sessionID, "protov", ProtoVersions)
	// return metadata.NewOutgoingContext(context.Background(), md)
}

func buildEnv(env *common.SessionEnv) *Env {
	protoEnv := Env{Url: env.URL, Headers: *env.Headers}
	if env.ConnectionState != nil {
		protoEnv.Cstate = *env.ConnectionState
	}
	return &protoEnv
}

func buildDisconnectEnv(env *common.SessionEnv) *Env {
	protoEnv := *buildEnv(env)

	if env.ChannelStates == nil {
		return &protoEnv
	}

	states := make(map[string]string)

	for id, state := range *env.ChannelStates {
		encodedState, _ := json.Marshal(state)

		states[id] = string(encodedState)
	}

	protoEnv.Istate = states

	return &protoEnv
}

func buildChannelEnv(id string, env *common.SessionEnv) *Env {
	protoEnv := *buildEnv(env)

	if env.ChannelStates == nil {
		return &protoEnv
	}

	if _, ok := (*env.ChannelStates)[id]; ok {
		protoEnv.Istate = (*env.ChannelStates)[id]
	}
	return &protoEnv
}

func (c *Controller) parseCommandResponse(r *CommandResponse, err error) (*common.CommandResult, error) {
	if err != nil {
		return nil, err
	}

	res := &common.CommandResult{
		Disconnect:     r.Disconnect,
		StopAllStreams: r.StopStreams,
		Streams:        r.Streams,
		StoppedStreams: r.StoppedStreams,
		Transmissions:  r.Transmissions,
	}

	if r.Env != nil {
		res.CState = r.Env.Cstate
		res.IState = r.Env.Istate
	}

	if r.Status.String() == "SUCCESS" {
		return res, nil
	}

	return res, fmt.Errorf("Application error: %s", r.ErrorMsg)
}
