package anycable

import (
	context "context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	grpc "google.golang.org/grpc"
)

type Connection interface {
	HandleOpen() error
	HandleCommand(identifier, command, data string) error
}

type ConnectionFactory func(c context.Context, env *Env, socket *Socket) (Connection, error)

// Server implements AnyCable server.
type Server struct {
	ConnectionFactory ConnectionFactory
}

// NewServer creates an instance of our server
func NewServer(cf ConnectionFactory) *Server {
	return &Server{
		ConnectionFactory: cf,
	}
}

func (s *Server) Serve(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	RegisterRPCServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

func (s *Server) Connect(c context.Context, r *ConnectionRequest) (*ConnectionResponse, error) {
	socket := Socket{}
	connection, err := s.ConnectionFactory(c, r.Env, &socket)
	if err != nil {
		return nil, err
	}
	if err := connection.HandleOpen(); err != nil {
		return &ConnectionResponse{
			Status:        Status_FAILURE,
			Transmissions: []string{`{"type":"disconnect","reason":"unauthorized","reconnect":false}`}, // TODO: Reason from err
		}, nil
	}

	return &ConnectionResponse{
		Status: Status_SUCCESS,
		// TODO: Identifiers
		Transmissions: []string{`{"type":"welcome"}`}, // TODO: User socket.
		// TODO: EnvResponse
	}, nil
}

type CommandResponseTransmission struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

func (t CommandResponseTransmission) Marshal() (string, error) {
	bs, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

type ResponseTransmission interface {
	Marshal() (string, error)
}

type Socket struct {
	transmissions []string
}

func (s *Socket) Write(t ResponseTransmission) error {
	json, err := t.Marshal()
	if err != nil {
		return err
	}
	s.transmissions = append(s.transmissions, json)
	return nil
}

type CommandSocket struct {
	r *CommandResponse
}

func (s *Server) Command(c context.Context, m *CommandMessage) (*CommandResponse, error) {
	socket := Socket{}
	connection, err := s.ConnectionFactory(c, m.Env, &socket)
	if err != nil {
		return nil, err
	}
	if err := connection.HandleCommand(m.Identifier, m.Command, m.Data); err != nil {
		return &CommandResponse{
			Status: Status_FAILURE,
			// TODO
		}, nil
	}
	return &CommandResponse{
		Status:        Status_SUCCESS,
		Transmissions: socket.transmissions,
	}, nil
}

func (s *Server) Disconnect(context.Context, *DisconnectRequest) (*DisconnectResponse, error) {
	return &DisconnectResponse{}, nil
}
