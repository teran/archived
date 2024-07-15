package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type testServer interface {
	DialContext(ctx context.Context) (*grpc.ClientConn, error)
	Run() error
	Close() error
	Server() *grpc.Server
}

type server struct {
	srv      *grpc.Server
	listener net.Listener
}

// New ...
func newTestServer(opts ...grpc.ServerOption) testServer {
	return &server{
		srv:      grpc.NewServer(opts...),
		listener: bufconn.Listen(2 * 1024 * 1024),
	}
}

// Close ...
func (s *server) Close() error {
	err := s.listener.Close()
	if err != nil {
		return err
	}

	s.srv.Stop()

	return nil
}

// DialContext ...
func (s *server) DialContext(ctx context.Context) (*grpc.ClientConn, error) {
	return grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(s.dial),
		grpc.WithInsecure(),
	)
}

func (s *server) dial(context.Context, string) (net.Conn, error) {
	return s.listener.(*bufconn.Listener).Dial()
}

// Run ...
func (s *server) Run() error {
	go func() {
		_ = s.srv.Serve(s.listener)
	}()

	return nil
}

func (s *server) Server() *grpc.Server {
	return s.srv
}
