package rpc

import (
	"errors"
	"github.com/riimi/tutorial-grpc-chat/clean/domain"
	"sync"
)

type Session struct {
	sync.RWMutex
	output chan *domain.Message
	user   *domain.User
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

func (s *Session) writeMessage(msg *domain.Message) error {
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

func (s *Session) ID() string {
	return s.user.Token
}
