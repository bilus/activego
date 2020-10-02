package anycable

import (
	context "context"
	"fmt"
	"log"
	"net"

	grpc "google.golang.org/grpc"
)

// Server implements AnyCable server.
type Server struct {
}

// type Hub struct {
// 	func Add(socketHandle interface{}, channelID)
// }

// NewServer creates an instance of our server
func NewServer() *Server {
	return &Server{}
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

func (s *Server) Connect(context.Context, *ConnectionRequest) (*ConnectionResponse, error) {
	return &ConnectionResponse{
		Status:        Status_SUCCESS,
		Transmissions: []string{`{"type":"welcome"}`},
	}, nil
}

func (s *Server) Command(context.Context, *CommandMessage) (*CommandResponse, error) {
	return &CommandResponse{}, nil
}

func (s *Server) Disconnect(context.Context, *DisconnectRequest) (*DisconnectResponse, error) {
	return &DisconnectResponse{}, nil
}
