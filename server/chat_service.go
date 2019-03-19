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
)

type Subscription interface {
	Updates() <-chan *domain.Message
	Close()
}

type Server interface {
	Broadcast(msg *domain.Message) error
	Subscribe(string) (Subscription, error)
	GenerateRandomToken(int) string
	RegisterUser(*domain.User) error
	UnregisterUser(*domain.User) error
	DuplicateToken(string) bool
	UserByToken(string) (*domain.User, error)
	ErrorLog(*domain.User, error, string)
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
	var token string
	for {
		token = s.server.GenerateRandomToken(16)
		if !s.server.DuplicateToken(token) {
			break
		}
	}
	newUser := &domain.User{
		Name:  req.Name,
		Token: token,
	}

	if err := s.server.RegisterUser(newUser); err != nil {
		return nil, err
	}

	return &protocol.LoginResponse{Token: newUser.Token}, nil
}

func (s *ChatService) Logout(ctx context.Context, req *protocol.LogoutRequest) (*protocol.LogoutResponse, error) {
	delUser, err := s.server.UserByToken(req.Token)
	if err != nil {
		st := status.New(codes.Unauthenticated, "invalid token")
		return nil, st.Err()
	}

	if err := s.server.UnregisterUser(delUser); err != nil {
		return nil, err
	}

	return &protocol.LogoutResponse{}, nil
}

func (s *ChatService) Subscribe(server protocol.ChatService_SubscribeServer) error {
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

func (s *ChatService) writePump(user *domain.User, server protocol.ChatService_SubscribeServer, sub Subscription) {
	for msg := range sub.Updates() {
		if err := server.Send(MsgToProtobuf(msg)); err != nil {
			s.server.ErrorLog(user, err, "")
		}
	}
}

func MsgToProtobuf(msg *domain.Message) *protocol.Message {
	if msg == nil {
		return nil
	}

	ret := &protocol.Message{
		Timestamp: ptypes.TimestampNow(),
	}
	switch msg.Type {
	case domain.MSGTYPE_LOGIN:
		ret.Event = &protocol.Message_Login_{
			Login: &protocol.Message_Login{
				Name: msg.Name,
			},
		}
	case domain.MSGTYPE_LOGOUT:
		ret.Event = &protocol.Message_Logout_{
			Logout: &protocol.Message_Logout{
				Name: msg.Name,
			},
		}
	case domain.MSGTYPE_MESSAGE:
		ret.Event = &protocol.Message_Message_{
			Message: &protocol.Message_Message{
				Name:    msg.Name,
				Message: msg.Text,
			},
		}
	case domain.MSGTYPE_SHUTDOWN:
		ret.Event = &protocol.Message_Shutdown_{
			Shutdown: &protocol.Message_Shutdown{},
		}
	}
	return ret
}

func protobufToMsg(msg *protocol.Message) *domain.Message {
	if msg == nil {
		return nil
	}

	switch msg.Event.(type) {
	case *protocol.Message_Login_:
		return &domain.Message{
			Type: domain.MSGTYPE_LOGIN,
			Name: msg.GetLogin().Name,
		}
	case *protocol.Message_Logout_:
		return &domain.Message{
			Type: domain.MSGTYPE_LOGOUT,
			Name: msg.GetLogout().Name,
		}
	case *protocol.Message_Message_:
		return &domain.Message{
			Type: domain.MSGTYPE_MESSAGE,
			Name: msg.GetMessage().Name,
			Text: msg.GetMessage().Message,
		}
	case *protocol.Message_Shutdown_:
		return &domain.Message{
			Type: domain.MSGTYPE_SHUTDOWN,
		}
	}
	return &domain.Message{
		Type: domain.MSGTYPE_UNKNOWN,
	}
}
