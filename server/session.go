package main

import (
	"errors"
	"sync"
	"tutorial-grpc-chat/pb"
)

type Session struct {
	sync.RWMutex
	output chan *pb.Message
	sync   chan interface{}
	stream pb.ChatService_SubscribeServer
	Id     string
	open   bool
	app    *ChatServer
}

var (
	ErrAlreadyClosed   = errors.New("[session] closed session")
	ErrWriteBufferFull = errors.New("[session] write buffer is full")
)

func (s *Session) closed() bool {
	s.RLock()
	defer s.RUnlock()
	return !s.open
}

func (s *Session) writeMessage(msg *pb.Message) error {
	if s.closed() {
		return ErrAlreadyClosed
	}

	select {
	case s.output <- msg:
	default:
		return ErrWriteBufferFull
	}
	return nil
}

func (s *Session) close() {
	if !s.closed() {
		s.Lock()
		s.open = false
		close(s.output)
		s.Unlock()
	}
}

func (s *Session) writePump() error {
	for {
		select {
		case msg, more := <-s.output:
			if !more {
				s.app.LogHandler(s, "[session] writepump channel closed")
				return nil
			}
			if err := s.stream.Send(msg); err != nil {
				s.app.ErrorHandler(s, err)
				return err
			}
		}
	}
}
