package grpc

import "context"

type Server struct{}
type Issuer interface {
	Sign(context.Context, string) (string, error)
}
type Revocation interface {
	Check(context.Context, string) (string, error)
}
type Replication interface {
	Sync(context.Context, []byte) error
}
type Challenge interface {
	Validate(context.Context, string) (bool, error)
}

func NewServer() *Server { return &Server{} }
