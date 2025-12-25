package main

type Server struct {
	host string
	port int
}
type Option func(*Server)

func withHost(host string) Option {
	return func(s *Server) {
		s.host = host
	}
}
func withPort(port int) Option {
	return func(s *Server) {
		s.port = port
	}
}

func NewServer(options ...Option) *Server {
	s := &Server{
		host: "127.0.0.1",
		port: 80,
	}
	for _, option := range options {
		option(s)
	}
	return s
}
