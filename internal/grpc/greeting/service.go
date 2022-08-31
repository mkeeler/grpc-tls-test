package greeting

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/mkeeler/grpc-tls-test/internal/proto/greeting"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

var _ greeting.GreetingServiceServer = (*Server)(nil)

func (s *Server) Register(grpcServer *grpc.Server) {
	greeting.RegisterGreetingServiceServer(grpcServer, s)
}

func (s *Server) Hello(_ context.Context, req *greeting.HelloRequest) (*greeting.HelloResponse, error) {
	name := req.GetName()
	if name == "" {
		name = "World"
	}

	return &greeting.HelloResponse{
		Greeting: fmt.Sprintf("Hello %s!", name),
	}, nil
}
