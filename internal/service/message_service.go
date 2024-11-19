package service

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"github.com/chatApp/internal/domain"
)

type MessageServiceImpl struct {
	mr domain.MessageRepository
	tr domain.Transactioner
}

func NewMessageService(mr domain.MessageRepository, tr domain.Transactioner) domain.MessageService {
	return &MessageServiceImpl{
		mr: mr,
		tr: tr,
	}

}

// FindAll implements domain.MessageService.
func (s *MessageServiceImpl) FindAll() (result []domain.Message, err error) {
	return s.mr.FindAll(context.Background())
}

// FindByID implements domain.MessageService.
func (s *MessageServiceImpl) FindByID(id uuid.UUID) (result domain.Message, err error) {
	return s.mr.FindByID(context.Background(), id)
}

// FindByReceiverID implements domain.MessageService.
func (s *MessageServiceImpl) FindByReceiverID(receiverID uuid.UUID) (result []domain.Message, err error) {
	return s.mr.FindByReceiverID(context.Background(), receiverID)
}

// FindBySenderID implements domain.MessageService.
func (s *MessageServiceImpl) FindBySenderID(senderID uuid.UUID) (result []domain.Message, err error) {
	return s.mr.FindBySenderID(context.Background(), senderID)
}

// Create implements domain.MessageService.
func (s *MessageServiceImpl) Create(in domain.CreateMessageInput) (result domain.Message, err error) {
	result = domain.Message{
		SenderID:   in.SenderID,
		ReceiverID: in.ReceiverID,
		Content:    in.Content,
	}
	ctx := context.Background()
	ctx, err = s.tr.Begin(ctx)
	if err != nil {
		return result, err
	}

	defer func() {
		s.tr.Rollback(ctx, err)
	}()
	err = s.mr.Create(ctx, &result)
	if err != nil {
		return result, err
	}
	mStatus := domain.MessageStatus{
		MessageID: result.ID,
		Mstatus:   domain.MstatusSent,
	}
	err = s.mr.CreateMessageStatus(ctx, &mStatus)
	if err != nil {
		return result, err
	}
	err = s.tr.Commit(ctx)
	if err != nil {
		return result, err
	}
	return result, nil

}

// Delete implements domain.MessageService.
func (s *MessageServiceImpl) Delete(id uuid.UUID) (err error) {
	return s.mr.Delete(context.Background(), id)
}

// UpdateMessageStatus implements domain.MessageService.
func (s *MessageServiceImpl) UpdateMessageStatus(in domain.UpdateMessageStatusInput) (err error) {
	mStatus := domain.MessageStatus{
		Mstatus:   in.Mstatus,
		MessageID: in.MessageID,
	}

	err = s.mr.UpdateMessageStatus(context.Background(), &mStatus)
	if err != nil {
		return err
	}
	return nil
}

// Update implements domain.MessageService.
func (s *MessageServiceImpl) Update(id uuid.UUID, in domain.UpdateMessageInput) (result domain.Message, err error) {
	ctx := context.Background()
	ctx, err = s.tr.Begin(ctx)
	if err != nil {
		return result, err
	}

	defer func() {
		s.tr.Rollback(ctx, err)
	}()
	result, err = s.mr.FindByID(ctx, id)
	if err != nil {
		return result, err
	}
	if in.Content != "" {
		result.Content = in.Content
	}
	err = s.mr.Update(ctx, &result)
	if err != nil {
		return result, err
	}
	err = s.tr.Commit(ctx)
	if err != nil {
		return result, err
	}
	return result, nil

}
