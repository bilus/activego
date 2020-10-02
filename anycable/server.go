package anycable

import (
	context "context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"

	grpc "google.golang.org/grpc"
)

// Server implements AnyCable server.
type Server struct {
	Connection Connection
}

// type Hub struct {
// 	func Add(socketHandle interface{}, channelID)
// }

type Connection interface {
	HandleOpen(env *Env) error
}

type TestConnection struct {
}

func request(env *Env) *http.Request {
	header := http.Header{}
	for key, value := range env.Headers {
		header.Set(key, value)
	}
	return &http.Request{Header: header}
}

func (c *TestConnection) HandleOpen(env *Env) error {
	testCases := map[string]func(env *Env) bool{
		"request_url": func(env *Env) bool {
			ok, err := regexp.MatchString("test=request_url", env.Url)
			return ok && err == nil
		},
		"cookies": func(env *Env) bool {
			username, err := request(env).Cookie("username")
			if err != nil {
				log.Printf("Error reading username from cookies: %v", err)
				return false
			}
			return username.Value == "john green"

		},
		"headers": func(env *Env) bool {
			return request(env).Header.Get("X-Api-Token") == "abc"
		},
	}
	u, err := url.Parse(env.Url)
	if err != nil {
		log.Printf("Failed parsing URL")
		return err
	}
	testName := u.Query().Get("test")
	if testName == "" {
		return nil
	}
	test, ok := testCases[testName]
	if !ok {
		log.Printf("No such test: %q", testName)
		return fmt.Errorf("no such test: %q", testName)
	}
	if success := test(env); !success {
		log.Printf("Test %q failed", testName)
		return fmt.Errorf("test %q failed", testName)
	}

	return nil
}

// NewServer creates an instance of our server
func NewServer(c Connection) *Server {
	return &Server{
		Connection: c,
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
	log.Println("Request header", r.Headers)
	err := s.Connection.HandleOpen(r.Env)
	if err != nil {
		return &ConnectionResponse{
			Status:   Status_ERROR,
			ErrorMsg: err.Error(),
		}, nil
	}

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
