package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/riimi/tutorial-grpc-chat/clean/domain"
	"github.com/riimi/tutorial-grpc-chat/server/protocol"
	"google.golang.org/grpc"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
)

//go:generate protoc -I./protocol -I/usr/local/include -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:./protocol --grpc-gateway_out=logtostderr=true,grpc_api_configuration=./protocol/gopher_service.yaml:./protocol --swagger_out=logtostderr=true,grpc_api_configuration=./protocol/gopher_service.yaml:./protocol --csharp_out=./protocol --grpc_out=./protocol --plugin=protoc-gen-grpc=/root/.nuget/packages/grpc.tools/1.19.0/tools/linux_x64/grpc_csharp_plugin gopher.proto

type ChatServer struct {
	Addr      string
	Gophers   map[string]*Session
	m         sync.RWMutex
	broadcast chan *protocol.Movement

	ErrorHandler func(*domain.User, error, string)
	LogHandler   func(*domain.User, string, ...interface{})
}

var (
	ErrNotValidSession     = errors.New("not valid session")
	ErrInvalidToken        = errors.New("invalid token")
	ErrBroadcastBufferFull = errors.New("broadcast buffer full")
)

func NewServer(addr string) *ChatServer {
	server := &ChatServer{
		Addr:      addr,
		Gophers:   make(map[string]*Session),
		broadcast: make(chan *protocol.Movement, 1024*16),

		ErrorHandler: func(*domain.User, error, string) {},
		LogHandler:   func(*domain.User, string, ...interface{}) {},
	}

	return server
}

func (s *ChatServer) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var opt []grpc.ServerOption
	srv := grpc.NewServer(opt...)
	protocol.RegisterGopherServiceServer(srv, NewChatService(s))

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

	if err := s.Broadcast(&domain.Message{
		Type: domain.MSGTYPE_SHUTDOWN,
		Name: "SYSTEM",
		Text: "server shutdown",
	}); err != nil {
		s.ErrorHandler(nil, err, "")
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
			s.m.RLock()
			for _, sess := range s.Gophers {
				/*
					if err := sess.writeMessage(msg); err != nil {
						s.ErrorHandler(sess.user, err, "")
					}
				*/
				sess.writeMessage(msg)
			}
			s.m.RUnlock()
			//s.LogHandler(nil, fmt.Sprintf("[broadcast] %v", *msg))

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
		output: make(chan *protocol.Movement, 1024*8),
	}

	s.m.Lock()
	defer s.m.Unlock()
	s.Gophers[user.Token] = sess

	if err := s.Broadcast(&domain.Message{
		Type: domain.MSGTYPE_LOGIN,
		Name: sess.user.Name,
	}); err != nil {
		s.ErrorHandler(user, err, "")
	}

	s.LogHandler(sess.user, "New Gohper!")
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
	sess.Close()

	if err := s.Broadcast(&domain.Message{
		Type: domain.MSGTYPE_LOGOUT,
		Name: sess.user.Name,
	}); err != nil {
		s.ErrorHandler(user, err, "")
	}
	s.LogHandler(sess.user, "Bye Gopher..")
	return nil
}

func (s *ChatServer) Subscribe(token string) (Subscription, error) {
	return s.SessionByID(token)
}

func (s *ChatServer) Broadcast(msg *domain.Message) error {
	select {
	case s.broadcast <- msg:
	default:
		return ErrBroadcastBufferFull
	}
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

func (s *ChatServer) ErrorLog(u *domain.User, err error, str string) {
	s.ErrorHandler(u, err, str)
}
