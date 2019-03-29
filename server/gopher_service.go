package main

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"github.com/riimi/tutorial-grpc-chat/server/domain"
	"github.com/riimi/tutorial-grpc-chat/server/protocol"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"math/rand"
)

type Subscription interface {
	Updates() <-chan *domain.GopherInfo
	Close()
}

type Server interface {
	Broadcast(msg *domain.GopherInfo) error
	Subscribe(string) (Subscription, error)
	GenerateRandomToken(int) string
	RegisterUser(*domain.GopherInfo) error
	UnregisterUser(*domain.GopherInfo) error
	DuplicateToken(string) bool
	UserByToken(string) (*domain.GopherInfo, error)
	ErrorLog(*domain.GopherInfo, error, string)
}

type ChatService struct {
	server Server
}

func NewChatService(s Server) *ChatService {
	return &ChatService{
		server: s,
	}
}

func (s *ChatService) Login(ctx context.Context, req *protocol.LoginRequest) (*protocol.LoginResponse, error) {
	var name string
	for {
		name = s.server.GenerateRandomToken(16)
		if !s.server.DuplicateToken(name) {
			break
		}
	}
	newUser := &domain.GopherInfo{
		Name: name,
		X:    rand.Int() % 500,
		Y:    rand.Int() % 500,
	}

	if err := s.server.RegisterUser(newUser); err != nil {
		return nil, err
	}

	return &protocol.LoginResponse{Name: newUser.Name}, nil
}

func (s *ChatService) Logout(ctx context.Context, req *protocol.LogoutRequest) (*protocol.LogoutResponse, error) {
	delUser, err := s.server.UserByToken(req.Name)
	if err != nil {
		st := status.New(codes.Unauthenticated, "invalid token")
		return nil, st.Err()
	}

	if err := s.server.UnregisterUser(delUser); err != nil {
		return nil, err
	}

	return &protocol.LogoutResponse{}, nil
}

func (s *ChatService) Move(server protocol.GopherService_MoveServer) error {
	token, ok := s.extractToken(server.Context())
	if !ok {
		st := status.New(codes.InvalidArgument, "missing token")
		return st.Err()
	}

	user, err := s.server.UserByToken(token)
	if err != nil {
		st := status.New(codes.Unauthenticated, "invalid token")
		return st.Err()
	}

	sub, err := s.server.Subscribe(token)
	if err != nil {
		st := status.New(codes.Unauthenticated, "invalid token")
		return st.Err()
	}

	go func() {
		for {
			in, err := server.Recv()
			if err == io.EOF {
				sub.Close()
				return
			} else if err != nil {
				s.server.ErrorLog(user, err, token)
				s.server.UnregisterUser(user)
				return
			}
			if err := s.server.Broadcast(protobufToMsg(in)); err != nil {
				s.server.ErrorLog(user, err, token)
				return
			}
		}
	}()

	s.writePump(user, server, sub)
	return nil
}

func (s *ChatService) extractToken(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md["myheader-token"]) == 0 {
		return "", false
	}
	return md["myheader-token"][0], true
}

func (s *ChatService) writePump(user *domain.GopherInfo, server protocol.GopherService_MoveServer, sub Subscription) {
	for msg := range sub.Updates() {
		if err := server.Send(MsgToProtobuf(msg)); err != nil {
			s.server.ErrorLog(user, err, "")
		}
	}
}
