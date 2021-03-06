package activego

import (
	context "context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/bilus/activego/anycable"
	"github.com/davecgh/go-spew/spew"
	grpc "google.golang.org/grpc"
)

type ActionData map[string]interface{}

type ChannelIdentifier struct {
	Channel string `json:"channel"`
	Params  map[string]interface{}
}

func (identifier *ChannelIdentifier) Unmarshal(bs []byte) error {
	params := make(map[string]interface{})
	if err := json.Unmarshal(bs, &params); err != nil {
		return err
	}
	i, ok := params["channel"]
	if !ok {
		return fmt.Errorf("missing %q in identifier", "channel")
	}
	identifier.Channel, ok = i.(string)
	if !ok {
		return fmt.Errorf("missing %q in identifier", "channel")
	}
	identifier.Params = params
	return nil
}

type ConnectionIdentifiers map[string]interface{}

func (c ConnectionIdentifiers) ToJSON() (string, error) {
	bs, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func (c *ConnectionIdentifiers) FromJSON(js string) error {
	return json.Unmarshal([]byte(js), c)
}

type Channel interface {
	HandleSubscribe() error
	HandleUnsubscribe() error
	HandleAction(action string, data ActionData) error
	IdentifierJSON() string
	// TODO: Params() and Channel() string
	Identifier() ChannelIdentifier
	StreamFrom(broadcasting string) error
	StopStreamFrom(broadcasting string) error
	Broadcast(stream string, data interface{}) error
	State() State
	Param(k string) interface{}
	Reject() error
}

// TODO: Pass ChannelIdentifier.
type ChannelFactory func(
	connection Connection,
	identifierJSON string,
	socket *Socket,
	broadcaster *Broadcaster) (Channel, error)

type Connection interface {
	HandleOpen() error
	HandleCommand(identifier, command, data string) error
	HandleClose(subscriptions []string) error
	Identifiers() ConnectionIdentifiers
	IdentifiedBy(key string, value interface{}) error
	State() State
	URL() *url.URL
	Header() http.Header
	Cookie(name string) (*http.Cookie, error)
	SaveToConnectionResponse(r *anycable.ConnectionResponse) error
	SaveToCommandResponse(r *anycable.CommandResponse) error
	Transmit(data interface{}) error
}

type ConnectionFactory func(
	c context.Context,
	env *anycable.Env,
	socket *Socket,
	broadcaster *Broadcaster,
	channelFactory ChannelFactory,
	identifiers ConnectionIdentifiers) (Connection, error)

// Server implements AnyCable server.
type Server struct {
	ConnectionFactory ConnectionFactory
	ChannelFactory    ChannelFactory
	Broadcaster       *Broadcaster
}

// NewServer creates an instance of our server
func NewServer(connectionFactory ConnectionFactory, channelFactory ChannelFactory, broadcaster *Broadcaster) *Server {
	return &Server{
		ConnectionFactory: connectionFactory,
		ChannelFactory:    channelFactory,
		Broadcaster:       broadcaster,
	}
}

func (s *Server) SetBroadcaster(broadcaster *Broadcaster) {
	s.Broadcaster = broadcaster
}

func (s *Server) Serve(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	anycable.RegisterRPCServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

func (s *Server) Connect(c context.Context, r *anycable.ConnectionRequest) (*anycable.ConnectionResponse, error) {
	fmt.Println("Connect")
	spew.Dump(*r)
	socket, err := NewSocket(r.Env, false)
	if err != nil {
		return nil, err
	}
	// TODO: Just pass new channel, not factory?
	connection, err := s.ConnectionFactory(c, r.Env, socket, s.Broadcaster, s.ChannelFactory, nil)
	if err != nil {
		return nil, err
	}
	var response anycable.ConnectionResponse
	if err := connection.HandleOpen(); err != nil {
		socket.Write(DisconnectResponseTransmission{
			Type:      "disconnect",
			Reason:    err.Error(),
			Reconnect: false,
		})
		response = anycable.ConnectionResponse{
			Status: anycable.Status_FAILURE,
		}
	} else {
		identifiersJSON, err := connection.Identifiers().ToJSON()
		if err != nil {
			return nil, err // TODO: Do we return err or anycable.ConnectionResponse + always nil?
		}
		response = anycable.ConnectionResponse{
			Status:      anycable.Status_SUCCESS,
			Identifiers: identifiersJSON,
			// TODO: EnvResponse
		}
	}
	if err := connection.SaveToConnectionResponse(&response); err != nil {
		return nil, err
	}
	fmt.Println("Response")
	spew.Dump(response)
	return &response, nil
}

func (s *Server) Command(c context.Context, m *anycable.CommandMessage) (*anycable.CommandResponse, error) {
	fmt.Println("Cmmand")
	spew.Dump(*m)
	socket, err := NewSocket(m.Env, false)
	if err != nil {
		return nil, err
	}

	identifiers := ConnectionIdentifiers{}
	if err := identifiers.FromJSON(m.ConnectionIdentifiers); err != nil {
		return nil, err
	}
	connection, err := s.ConnectionFactory(c, m.Env, socket, s.Broadcaster, s.ChannelFactory, identifiers)
	if err != nil {
		return nil, err
	}
	var response anycable.CommandResponse
	if err := connection.HandleCommand(m.Identifier, m.Command, m.Data); err != nil {
		response = anycable.CommandResponse{
			Status:   anycable.Status_FAILURE,
			ErrorMsg: fmt.Sprintf("Error handling command %q: %v", m.Command, err),
			// TODO
		}
	} else {
		response = anycable.CommandResponse{
			Status: anycable.Status_SUCCESS,
		}
	}
	if err := connection.SaveToCommandResponse(&response); err != nil {
		return nil, err
	}
	fmt.Println("Response")
	spew.Dump(response)
	return &response, nil
}

func (s *Server) Disconnect(c context.Context, r *anycable.DisconnectRequest) (*anycable.DisconnectResponse, error) {
	fmt.Println("Disconnect")
	spew.Dump(*r)
	socket, err := NewSocket(r.Env, true)
	if err != nil {
		return nil, err
	}
	identifiers := ConnectionIdentifiers{}
	if err := identifiers.FromJSON(r.Identifiers); err != nil {
		return nil, err
	}
	connection, err := s.ConnectionFactory(c, r.Env, socket, s.Broadcaster, s.ChannelFactory, identifiers)
	if err != nil {
		return nil, err
	}
	var response anycable.DisconnectResponse
	if err := connection.HandleClose(r.Subscriptions); err != nil {
		response = anycable.DisconnectResponse{
			Status:   anycable.Status_FAILURE,
			ErrorMsg: fmt.Sprintf("Error handling disconnect: %v", err),
		}
	} else {
		// TODO: Is DisconnectResponseTransmission the best name?
		err = s.Broadcaster.Broadcast(r.Identifiers, DisconnectResponseTransmission{
			Type:      "disconnect",
			Reason:    "remote",
			Reconnect: true,
		})
		if err != nil {
			log.Printf("Error broadcasting disconnect command: %v", err)
		}
		response = anycable.DisconnectResponse{
			Status: anycable.Status_SUCCESS,
		}
	}
	fmt.Println("Response")
	spew.Dump(response)
	return &response, nil
}
