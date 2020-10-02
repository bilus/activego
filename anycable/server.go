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

func request(env *Env) (*http.Request, error) {
	header := http.Header{}
	for key, value := range env.Headers {
		header.Set(key, value)
	}
	u, err := url.Parse(env.Url)
	if err != nil {
		return nil, err
	}
	return &http.Request{Header: header, URL: u}, nil
}

func (c *TestConnection) HandleOpen(env *Env) error {
	testCases := map[string]func(req *http.Request) bool{
		"request_url": func(req *http.Request) bool {
			ok, err := regexp.MatchString("test=request_url", req.URL.String())
			return ok && err == nil
		},
		"cookies": func(req *http.Request) bool {
			username, err := req.Cookie("username")
			if err != nil {
				log.Printf("Error reading username from cookies: %v", err)
				return false
			}
			return username.Value == "john green"

		},
		"headers": func(req *http.Request) bool {
			return req.Header.Get("X-Api-Token") == "abc"
		},
		"reasons": func(req *http.Request) bool {
			return req.URL.Query().Get("reason") != "unauthorized"
		},
	}

	req, err := request(env)
	if err != nil {
		return err
	}
	testName := req.URL.Query().Get("test")
	if testName == "" {
		return nil
	}
	test, ok := testCases[testName]
	if !ok {
		log.Printf("No such test: %q", testName)
		return fmt.Errorf("no such test: %q", testName)
	}
	if success := test(req); !success {
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
			Status:        Status_FAILURE,
			Transmissions: []string{`{"type":"disconnect","reason":"unauthorized","reconnect":false}`}, // TDO
		}, nil
	}

	return &ConnectionResponse{
		Status: Status_SUCCESS,
		// TODO: Identifiers
		Transmissions: []string{`{"type":"welcome"}`},
		// TODO: EnvResponse
	}, nil
}

func (s *Server) Command(context.Context, *CommandMessage) (*CommandResponse, error) {
	return &CommandResponse{}, nil
}

func (s *Server) Disconnect(context.Context, *DisconnectRequest) (*DisconnectResponse, error) {
	return &DisconnectResponse{}, nil
}
