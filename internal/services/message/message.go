package message

import (
	"context"
	"net/http"
	"time"

	"go-form-hub/internal/model"
	"go-form-hub/internal/repository"
	resp "go-form-hub/internal/services/service_response"

	validator "github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

type Service interface {
	MessageSave(ctx context.Context, message *model.MessageSave) (*resp.Response, error)
	MessageCheck(ctx context.Context) (*resp.Response, error)
	ChatList(ctx context.Context) (*resp.Response, error)
	ChatGet(ctx context.Context, id int64) (*resp.Response, error)
}

type messageService struct {
	messageRepository repository.MessageRepository
	sanitizer         *bluemonday.Policy
	validate          *validator.Validate
}

func NewMessageService(messageRepository repository.MessageRepository, validate *validator.Validate) Service {
	sanitizer := bluemonday.UGCPolicy()
	return &messageService{
		messageRepository: messageRepository,
		validate:          validate,
		sanitizer:         sanitizer,
	}
}

func (s *messageService) MessageSave(ctx context.Context, message *model.MessageSave) (*resp.Response, error) {
	currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

	message.SenderID = currentUser.ID
	message.SendAt = time.Now().UTC()

	err := s.messageRepository.Insert(ctx, message)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	return resp.NewResponse(http.StatusOK, nil), nil
}

func (s *messageService) MessageCheck(ctx context.Context) (*resp.Response, error) {
	currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

	result, err := s.messageRepository.CheckUnreadForUser(ctx, currentUser.ID)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	return resp.NewResponse(http.StatusOK, result), nil
}

func (s *messageService) ChatGet(ctx context.Context, id int64) (*resp.Response, error) {
	currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

	err := s.messageRepository.ReadAllInChat(ctx, currentUser.ID, id)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	result, err := s.messageRepository.GetChatByIDs(ctx, currentUser.ID, id)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	if result == nil {
		return resp.NewResponse(http.StatusNotFound, nil), nil
	}

	return resp.NewResponse(http.StatusOK, result), nil
}

func (s *messageService) ChatList(ctx context.Context) (*resp.Response, error) {
	currentUser := ctx.Value(model.ContextCurrentUser).(*model.UserGet)

	result, err := s.messageRepository.GetChatListByUserID(ctx, currentUser.ID)
	if err != nil {
		return resp.NewResponse(http.StatusInternalServerError, nil), err
	}

	chatList := &model.ChatList{
		Chats: result,
	}
	chatList.Count = len(result)

	return resp.NewResponse(http.StatusOK, chatList), nil
}
