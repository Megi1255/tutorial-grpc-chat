package chatserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/riimi/tutorial-grpc-chat/clean/domain"
	"github.com/riimi/tutorial-grpc-chat/clean/interface/rpc"
	"github.com/riimi/tutorial-grpc-chat/clean/interface/rpc/protocol"
	"github.com/riimi/tutorial-grpc-chat/clean/usecase"
	"google.golang.org/grpc"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
)

//go:generate protoc -I../../interface/rpc/protocol -I/usr/local/include -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:../../interface/rpc/protocol --grpc-gateway_out=logtostderr=true:../../interface/rpc/protocol --swagger_out=logtostderr=true:../../interface/rpc/protocol chat-gateway.proto

type ChatServer struct {
	Addr      string
	Gophers   map[string]*Session
	m         sync.RWMutex
	broadcast chan *domain.Message

	ErrorHandler func(*Session, error, string)
	LogHandler   func(*Session, string, ...interface{})
}

var (
	ErrNotValidSession = errors.New("not valid session")
	ErrInvalidToken    = errors.New("invalid token")
)

func NewServer(addr string) *ChatServer {
	server := &ChatServer{
		Addr:      addr,
		Gophers:   make(map[string]*Session),
		broadcast: make(chan *domain.Message, 100),

		ErrorHandler: func(*Session, error, string) {},
		LogHandler:   func(*Session, string, ...interface{}) {},
	}

	return server
}

func (s *ChatServer) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var opt []grpc.ServerOption
	srv := grpc.NewServer(opt...)
	protocol.RegisterChatServiceServer(srv, rpc.NewChatService(
		usecase.NewChatUsecase(s),
		func(token string, err error) {
			s.ErrorHandler(nil, err, token)
		},
	))

	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		s.ErrorHandler(nil, err, "[main] failed to listen")
		return err
	}

	go s.Hub(ctx)
	go func() {
		s.LogHandler(nil, fmt.Sprintf("[main] server is running: %s", s.Addr))
		srv.Serve(lis)
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	cancel()

	s.broadcast <- &domain.Message{
		Type: domain.MSGTYPE_SHUTDOWN,
		Name: "SYSTEM",
		Text: "server shutdown",
	}

	s.LogHandler(nil, "[main] server closed")
	srv.GracefulStop()
	close(s.broadcast)
	return nil
}

func (s *ChatServer) Hub(ctx context.Context) {
	for {
		select {
		case msg := <-s.broadcast:
			for _, sess := range s.Gophers {
				if err := sess.writeMessage(msg); err != nil {
					s.ErrorHandler(sess, err, "")
				}
			}
			s.LogHandler(nil, fmt.Sprintf("[broadcast] %v", *msg))
		case <-ctx.Done():
			return
		}
	}
}

func (s *ChatServer) GenerateRandomToken(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for {
		for i := range b {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
		return string(b)
	}
}

func (s *ChatServer) DuplicateToken(token string) bool {
	_, err := s.SessionByID(token)
	return err == nil
}

func (s *ChatServer) RegisterUser(user *domain.User) error {
	sess := &Session{
		open:   true,
		app:    s,
		user:   user,
		output: make(chan *domain.Message, 100),
	}

	s.m.Lock()
	defer s.m.Unlock()
	s.Gophers[user.Token] = sess

	s.broadcast <- &domain.Message{
		Type: domain.MSGTYPE_LOGIN,
	}

	s.LogHandler(sess, "New Gohper!")
	return nil
}

func (s *ChatServer) UnregisterUser(user *domain.User) error {
	sess, err := s.SessionByID(user.Token)
	if err != nil {
		return err
	}

	s.m.Lock()
	defer s.m.Unlock()
	user = sess.user
	delete(s.Gophers, sess.user.Token)
	sess.close()
	s.LogHandler(sess, "Bye Gopher..")
	return nil
}

func (s *ChatServer) Recv(token string) (<-chan *domain.Message, error) {
	sess, err := s.SessionByID(token)
	if err != nil {
		return nil, err
	}
	sess.open = true
	return sess.output, nil
}

func (s *ChatServer) Broadcast(msg *domain.Message) error {
	s.broadcast <- msg
	return nil
}

func (s *ChatServer) UserByToken(token string) (*domain.User, error) {
	sess, err := s.SessionByID(token)
	if err != nil {
		return nil, err
	}
	return sess.user, nil
}

func (s *ChatServer) SessionByID(id string) (*Session, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	sess, ok := s.Gophers[id]
	if !ok {
		return nil, ErrNotValidSession
	}
	return sess, nil
}
