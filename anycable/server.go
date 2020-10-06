package anycable

import (
	context "context"
	"fmt"
	"log"
	"net"

	"github.com/davecgh/go-spew/spew"
	grpc "google.golang.org/grpc"
)

type CommandData map[string]interface{}

type Channel interface {
	HandleSubscribe() error
	HandleUnsubscribe() error
	HandleAction(action string, data CommandData) error
	Identifier() string
}

type ChannelFactory func(identifier string, socket *Socket, broadcaster *Broadcaster) (Channel, error)

type Connection interface {
	HandleOpen() error
	HandleCommand(identifier, command, data string) error
	HandleClose(subscriptions []string) error
}

type ConnectionFactory func(context.Context, *Env, *Socket, *Broadcaster, ChannelFactory) (Connection, error)

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
	fmt.Println("Connect")
	spew.Dump(*r)
	socket := Socket{}
	// TODO: Just pass new channel, not factory?
	connection, err := s.ConnectionFactory(c, r.Env, &socket, s.Broadcaster, s.ChannelFactory)
	if err != nil {
		return nil, err
	}
	var response ConnectionResponse
	if err := connection.HandleOpen(); err != nil {
		response = ConnectionResponse{
			Status: Status_FAILURE,
		}
	} else {
		response = ConnectionResponse{
			Status: Status_SUCCESS,
			// TODO: Identifiers
			// TODO: EnvResponse
		}
	}
	socket.SaveToConnectionResponse(&response)
	fmt.Println("Response")
	spew.Dump(response)
	return &response, nil
}

func (s *Server) Command(c context.Context, m *CommandMessage) (*CommandResponse, error) {
	fmt.Println("Cmmand")
	spew.Dump(*m)
	socket := Socket{}
	connection, err := s.ConnectionFactory(c, m.Env, &socket, s.Broadcaster, s.ChannelFactory)
	if err != nil {
		return nil, err
	}
	var response CommandResponse
	if err := connection.HandleCommand(m.Identifier, m.Command, m.Data); err != nil {
		response = CommandResponse{
			Status:   Status_FAILURE,
			ErrorMsg: fmt.Sprintf("Error handling command %q: %v", m.Command, err),
			// TODO
		}
	} else {
		response = CommandResponse{
			Status: Status_SUCCESS,
		}
	}
	socket.SaveToCommandResponse(&response)
	fmt.Println("Response")
	spew.Dump(response)
	return &response, nil
}

func (s *Server) Disconnect(c context.Context, r *DisconnectRequest) (*DisconnectResponse, error) {
	fmt.Println("Disconnect")
	spew.Dump(*r)
	socket := Socket{}
	connection, err := s.ConnectionFactory(c, r.Env, &socket, s.Broadcaster, s.ChannelFactory)
	if err != nil {
		return nil, err
	}
	var response DisconnectResponse
	if err := connection.HandleClose(r.Subscriptions); err != nil {
		response = DisconnectResponse{
			Status:   Status_FAILURE,
			ErrorMsg: fmt.Sprintf("Error handling disconnect: %v", err),
		}
	} else {
		response = DisconnectResponse{
			Status: Status_SUCCESS,
		}
	}
	fmt.Println("Response")
	spew.Dump(response)
	return &response, nil
}
