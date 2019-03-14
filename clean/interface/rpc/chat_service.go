package rpc

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"github.com/riimi/tutorial-grpc-chat/clean/domain"
	"github.com/riimi/tutorial-grpc-chat/clean/interface/rpc/protocol"
	"github.com/riimi/tutorial-grpc-chat/clean/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
)

type ChatService struct {
	app         *ChatServer
	chatUsecase *usecase.ChatUsecase
}

func NewChatService(s *ChatServer, c *usecase.ChatUsecase) *ChatService {
	return &ChatService{
		app:         s,
		chatUsecase: c,
	}
}

func (s *ChatService) Login(ctx context.Context, req *protocol.LoginRequest) (*protocol.LoginResponse, error) {
	user, err := s.chatUsecase.Join(req.Name)
	if err != nil {
		return nil, err
	}

	return &protocol.LoginResponse{Token: user.Token}, nil
}

func (s *ChatService) Logout(ctx context.Context, req *protocol.LogoutRequest) (*protocol.LogoutResponse, error) {
	if _, err := s.chatUsecase.Quit(req.Token); err != nil {
		return nil, err
	}

	return &protocol.LogoutResponse{}, nil
}

func (s *ChatService) Subscribe(server protocol.ChatService_SubscribeServer) error {
	token, ok := s.extractToken(server.Context())
	if !ok {
		st := status.New(codes.Unauthenticated, "invalid token")
		return st.Err()
	}

	output, err := s.chatUsecase.Subscribe(token)
	if err != nil {
		return err
	}
	go s.writePump(token, server, output)

	for {
		in, err := server.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			s.app.ErrorHandler(nil, err, token)
			return err
		}
		if err := s.chatUsecase.SendMessageAll(protobufToMsg(in)); err != nil {
			s.app.ErrorHandler(nil, err, token)
			return err
		}
	}
}

func (s *ChatService) extractToken(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md["myheader-token"]) == 0 {
		return "", false
	}
	return md["myheader-token"][0], true
}

func (s *ChatService) writePump(token string, server protocol.ChatService_SubscribeServer, output <-chan *domain.Message) {
	for {
		select {
		case msg, more := <-output:
			if !more {
				return
			}
			if err := server.Send(MsgToProtobuf(msg)); err != nil {
				s.app.ErrorHandler(nil, err, token)
				return
			}
		}
	}
}

func MsgToProtobuf(msg *domain.Message) *protocol.Message {
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
