package usecase

import "github.com/riimi/tutorial-grpc-chat/clean/domain"

type ChatService interface {
	Broadcast(msg domain.Message) error
}

type ChatUsecase struct {
	ChatService ChatService
}

func (u *ChatUsecase) SendAll(mt domain.MessageType, senderName, message string) error {
	msg := domain.Message{
		Type: mt,
		Name: senderName,
		Text: message,
	}

	if err := u.ChatService.Broadcast(msg); err != nil {
		return err
	}

	return nil
}
