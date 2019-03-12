package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"sync"

	"github.com/riimi/tutorial-grpc-chat/pb"
)

//go:generate protoc -I../pb -I/usr/local/include -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:../pb --grpc-gateway_out=logtostderr=true:../pb --swagger_out=logtostderr=true:../pb chat-gateway.proto

type ChatServer struct {
	Ctx        context.Context
	Gophers    map[string]*Session
	m          sync.RWMutex
	Broadcast  chan *pb.Message
	Connect    chan *Session
	Disconnect chan *Session

	ErrorHandler func(*Session, error)
	LogHandler   func(*Session, string)
}

var (
	ErrNotValidSession = errors.New("[broadcast] not valid session")
)

func (s *ChatServer) Run(ctx context.Context) {
	for {
		select {
		case msg := <-s.Broadcast:
			sender, err := s.SessionByID(msg.Id)
			if err != nil {
				s.ErrorHandler(sender, err)
			}
			s.m.RLock()
			for _, sess := range s.Gophers {
				sess.writeMessage(msg)
			}
			s.m.RUnlock()
			s.LogHandler(sender, fmt.Sprintf("[broadcast] %v", *msg))
		case sess := <-s.Connect:
			s.LogHandler(sess, "[connect]")
			s.m.Lock()
			sess.Id = s.generateRandomId(16)
			s.Gophers[sess.Id] = sess
			s.m.Unlock()
			sess.sync <- sess.Id
		case sess := <-s.Disconnect:
			s.LogHandler(sess, "[disconnect]")
			if _, ok := s.Gophers[sess.Id]; ok {
				s.m.Lock()
				delete(s.Gophers, sess.Id)
				s.m.Unlock()
				sess.close()
			}
		case <-ctx.Done():
			s.LogHandler(nil, "[terminate]")
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
		if _, ok := s.Gophers[string(b)]; ok {
			continue
		}
		return string(b)
	}
}

func (s *ChatServer) Send(ctx context.Context, msg *pb.Message) (*empty.Empty, error) {
	s.Broadcast <- msg
	return &empty.Empty{}, nil
}

func (s *ChatServer) Subscribe(e *empty.Empty, stream pb.ChatService_SubscribeServer) error {
	sess := &Session{
		app:    s,
		output: make(chan *pb.Message, 32),
		sync:   make(chan interface{}),
		open:   true,
	}
	s.Connect <- sess
	defer func() {
		s.Disconnect <- sess
	}()
	<-sess.sync
	sess.stream = stream
	s.Broadcast <- &pb.Message{
		Id:   sess.Id,
		Text: "New Gopher!!",
	}

	return sess.writePump()
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

func main() {
	port := flag.Int("port", 40040, "port")
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("[main] failed to listen: %v", err)
	}

	var opt []grpc.ServerOption
	server := grpc.NewServer(opt...)
	gs := NewServer()
	gs.ErrorHandler = func(sess *Session, err error) {
		if sess != nil {
			log.Print(*sess)
		}
		log.Printf("%v\n", err)
	}
	gs.LogHandler = func(sess *Session, msg string) {
		if sess != nil {
			log.Printf("SessionId(%s) ", sess.Id)
		}
		log.Print(msg)
	}
	pb.RegisterChatServiceServer(server, gs)
	log.Printf("[main] server is running at port: %d", *port)
	log.Fatal(server.Serve(lis))
}

func NewServer() *ChatServer {
	server := &ChatServer{
		Gophers:    make(map[string]*Session),
		Broadcast:  make(chan *pb.Message, 100),
		Connect:    make(chan *Session, 100),
		Disconnect: make(chan *Session, 100),
		Ctx:        context.Background(),

		ErrorHandler: func(*Session, error) {},
		LogHandler:   func(*Session, string) {},
	}

	go server.Run(server.Ctx)

	return server
}
