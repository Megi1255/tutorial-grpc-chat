package rpc

import (
	"context"
	"github.com/riimi/tutorial-grpc-chat/clean/interface/rpc/protocol"
	"github.com/riimi/tutorial-grpc-chat/clean/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatService struct {
	userUsecase usecase.UserUsecase
	chatUsecase usecase.ChatUsecase
}

func NewChatService(u usecase.UserUsecase, c usecase.ChatUsecase) *ChatService {
	return &ChatService{
		userUsecase: u,
		chatUsecase: c,
	}
}

func (s *ChatService) Login(ctx context.Context, req *protocol.LoginRequest) (*protocol.LoginResponse, error) {
	c := Identifier{
		input: req.Name,
		ack:   make(chan Ack, 1),
	}
	s.Connect <- c
	ret := <-c.ack
	if ret.Err != nil {
		st := status.New(codes.Internal, "failed to connect")
		return nil, st.Err()
	}
	return &protocol.LoginResponse{Token: ret.Output}, nil
}

func (s *ChatService) Logout(ctx context.Context, req *protocol.LogoutRequest) (*protocol.LogoutResponse, error) {
	c := Identifier{
		input: req.Token,
		ack:   make(chan Ack, 1),
	}
	s.Disconnect <- c
	ret := <-c.ack
	if ret.Err != nil {
		st := status.New(codes.InvalidArgument, "invalid token")
		return nil, st.Err()
	}
	return &protocol.LogoutResponse{}, nil
}

func (s *ChatService) Subscribe(server protocol.ChatService_SubscribeServer) error {
	return nil
}
