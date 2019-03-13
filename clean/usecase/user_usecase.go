package usecase

import (
	"github.com/riimi/tutorial-grpc-chat/clean/domain"
)

type UserService interface {
	GenerateRandomToken(int) string
	DuplicateToken(string) bool
	RegisterUser(*domain.User) error
	UnregisterUser(*domain.User) error
}

type UserUsecase struct {
	UserService UserService
}

func (u *UserUsecase) Join(name string) (*domain.User, error) {
	var token string
	for {
		token = u.UserService.GenerateRandomToken(16)
		if !u.UserService.DuplicateToken(token) {
			break
		}
	}
	newUser := &domain.User{
		Name: name,
		Id:   token,
	}

	if err := u.UserService.RegisterUser(newUser); err != nil {
		return newUser, err
	}

	return newUser, nil
}

func (u *UserUsecase) Quit(token string) (*domain.User, error) {
	delUser := &domain.User{
		Id: token,
	}

	if err := u.UserService.UnregisterUser(delUser); err != nil {
		return delUser, err
	}

	return delUser, nil
}

func (u *UserUsecase) ListUser() ([]*domain.User, error) {
	return []*domain.User{}, nil
}
