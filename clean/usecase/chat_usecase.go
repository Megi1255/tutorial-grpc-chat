package usecase

import (
	"errors"
	"github.com/riimi/tutorial-grpc-chat/clean/domain"
)

type ChatServer interface {
	Broadcast(msg *domain.Message) error
	Recv(string) (<-chan *domain.Message, error)
	GenerateRandomToken(int) string
	RegisterUser(*domain.User) error
	UnregisterUser(*domain.User) error
	DuplicateToken(string) bool
	UserByToken(string) (*domain.User, error)
}

type ChatUsecase struct {
	server ChatServer
}

func NewChatUsecase(chatService ChatServer) *ChatUsecase {
	return &ChatUsecase{
		server: chatService,
	}
}

func (u *ChatUsecase) SendMessageAll(msg *domain.Message) error {
	return u.server.Broadcast(msg)
}

func (u *ChatUsecase) Subscribe(token string) (<-chan *domain.Message, error) {
	if !u.server.DuplicateToken(token) {
		return nil, errors.New("Unauthenticated")
	}

	return u.server.Recv(token)
}

func (u *ChatUsecase) Join(name string) (*domain.User, error) {
	var token string
	for {
		token = u.server.GenerateRandomToken(16)
		if !u.server.DuplicateToken(token) {
			break
		}
	}
	newUser := &domain.User{
		Name:  name,
		Token: token,
	}

	if err := u.server.RegisterUser(newUser); err != nil {
		return newUser, err
	}

	return newUser, nil
}

func (u *ChatUsecase) Quit(token string) (*domain.User, error) {
	delUser, err := u.server.UserByToken(token)
	if err != nil {
		return nil, err
	}

	if err := u.server.UnregisterUser(delUser); err != nil {
		return delUser, err
	}

	return delUser, nil
}

func (u *ChatUsecase) ListUser() ([]*domain.User, error) {
	return []*domain.User{}, nil
}
