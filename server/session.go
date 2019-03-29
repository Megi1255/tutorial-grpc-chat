package main

import (
	"errors"
	"github.com/riimi/tutorial-grpc-chat/server/domain"
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
	ErrAlreadyClosed   = errors.New("closed session")
	ErrWriteBufferFull = errors.New("write buffer is full")
	ErrMessageNil      = errors.New("message should be not nil")
)

func (s *Session) closed() bool {
	s.RLock()
	defer s.RUnlock()
	return !s.open
}

func (s *Session) writeMessage(msg *domain.Message) error {
	defer func() {
		if r := recover(); r != nil {
			s.app.LogHandler(s.user, "%v", r)
		}
	}()
	if msg == nil {
		return ErrMessageNil
	}
	if s.closed() {
		return ErrAlreadyClosed
	}
	s.RLock()
	defer s.RUnlock()

	select {
	case s.output <- msg:
	default:
		return ErrWriteBufferFull
	}
	return nil
}

func (s *Session) Close() {
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

func (s *Session) Updates() <-chan *domain.Message {
	return s.output
}
