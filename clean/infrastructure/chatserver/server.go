package chatserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/riimi/tutorial-grpc-chat/clean/interface/rpc"
	"github.com/riimi/tutorial-grpc-chat/clean/usecase"
	"google.golang.org/grpc"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/riimi/tutorial-grpc-chat/pb"
)

//go:generate protoc -I../pb -I/usr/local/include -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:../pb --grpc-gateway_out=logtostderr=true:../pb --swagger_out=logtostderr=true:../pb chat-gateway.proto

type ChatServer struct {
	Addr       string
	Ctx        context.Context
	Gophers    map[string]*Session
	m          sync.RWMutex
	Broadcast  chan *pb.Message
	Connect    chan Identifier
	Disconnect chan Identifier

	ErrorHandler func(*Session, error, string)
	LogHandler   func(*Session, string, ...interface{})
}

type Identifier struct {
	input string
	ack   chan Ack
}

type Ack struct {
	Output string
	Err    error
}

var (
	ErrNotValidSession = errors.New("not valid session")
	ErrInvalidToken    = errors.New("invalid token")
)

func NewServer(addr string) *ChatServer {
	server := &ChatServer{
		Addr:       addr,
		Gophers:    make(map[string]*Session),
		Broadcast:  make(chan *pb.Message, 100),
		Connect:    make(chan Identifier, 100),
		Disconnect: make(chan Identifier, 100),

		ErrorHandler: func(*Session, error, string) {},
		LogHandler:   func(*Session, string, ...interface{}) {},
	}

	go server.Hub(server.Ctx)

	return server
}

func (s *ChatServer) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var opt []grpc.ServerOption
	srv := grpc.NewServer(opt...)
	pb.RegisterChatServiceServer(srv, rpc.NewChatService(
		usecase.UserUsecase{UserService: s},
		usecase.ChatUsecase{ChatService: s},
	))

	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		s.ErrorHandler(nil, err, "[main] failed to listen")
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

	s.Broadcast <- &pb.Message{
		Timestamp: ptypes.TimestampNow(),
		Event: &pb.Message_Shutdown_{
			Shutdown: &pb.Message_Shutdown{},
		},
	}

	close(s.Broadcast)
	close(s.Connect)
	close(s.Disconnect)
	s.LogHandler(nil, "[main] server closed")
	srv.GracefulStop()
}

func (s *ChatServer) Hub(ctx context.Context) {
	for {
		select {
		case msg := <-s.Broadcast:
			for _, sess := range s.Gophers {
				sess.writeMessage(msg)
			}
			//s.LogHandler(sender, fmt.Sprintf("[broadcast] %v", *msg))
		case cont := <-s.Connect:
			sess := s.NewSession(cont.input)
			s.LogHandler(sess, "[connect]")
			cont.ack <- Ack{Output: sess.Id, Err: nil}
		case cont := <-s.Disconnect:
			sess, ok := s.DeleteSession(cont.input)
			if ok {
				s.LogHandler(sess, "[disconnect]")
				cont.ack <- Ack{Output: sess.name, Err: nil}
			} else {
				s.ErrorHandler(sess, ErrInvalidToken, "[disconnect]")
				cont.ack <- Ack{Output: "", Err: ErrInvalidToken}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *ChatServer) generateRandomId(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for {
		for i := range b {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
		return string(b)
	}
}

func (s *ChatServer) SessionByID(id string) (*Session, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	sess, ok := s.Gophers[id]
	if !ok {
		return &Session{Id: id}, ErrNotValidSession
	}
	return sess, nil
}

func (s *ChatServer) NewSession(name string) *Session {
	sess := &Session{
		app:    s,
		output: make(chan *pb.Message, 32),
		open:   true,
		name:   name,
	}

	for {
		id := s.generateRandomId(16)
		if _, ok := s.Gophers[id]; !ok {
			sess.Id = id
			s.Gophers[id] = sess
			break
		}
	}

	return sess
}

func (s *ChatServer) DeleteSession(id string) (*Session, bool) {
	sess, ok := s.Gophers[id]
	if ok {
		sess.close()
		delete(s.Gophers, id)
	}

	return sess, ok
}
