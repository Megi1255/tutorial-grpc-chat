package chatserver

import (
	"errors"
	"github.com/riimi/tutorial-grpc-chat/pb"
	"sync"
)

type Session struct {
	sync.RWMutex
	output chan *pb.Message
	stream pb.ChatService_SubscribeServer
	Id     string
	name   string
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
				s.app.ErrorHandler(s, err, "[session] writepump failed to send")
				return err
			}
		}
	}
}
