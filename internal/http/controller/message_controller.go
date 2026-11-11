package controller

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/http/transport"
	"github.com/Rizwank123/chatApp/internal/pkg/security"
	"github.com/Rizwank123/chatApp/internal/websockets"
)

type ChatController struct {
	cs  domain.ChatService
	hub *websockets.Hub
}

func NewChatController(cs domain.ChatService, hub *websockets.Hub) ChatController {
	return ChatController{
		cs:  cs,
		hub: hub,
	}
}

// SendFriendRequest sends a friend request to another user.
//
//	@Summary		Send a friend request
//	@Description	Send a friend request to a specific user
//	@Tags			Friends
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string							true	"Bearer"
//	@Param			body			body		domain.CreateFriendshipInput	true	"Friendship creation input"
//	@Success		201				{object}	domain.BaseResponse{data=domain.Friendship}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/friends [post]
func (c ChatController) SendFriendRequest(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}

	var input domain.CreateFriendshipInput
	if err := transport.DecodeAndValidateRequestBody(ctx, &input); err != nil {
		return err
	}

	result, err := c.cs.SendFriendRequest(personnelID, input)
	if err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusCreated, result)
}

// AcceptFriendRequest accepts a friend request.
//
//	@Summary		Accept a friend request
//	@Description	Accept a friend request from a specific user
//	@Tags			Friends
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer"
//	@Param			requestId		path		string	true	"Friend request ID"
//	@Success		200				{object}	domain.BaseResponse{data=nil}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/friends/{requestId}/accept [post]
func (c ChatController) AcceptFriendRequest(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}
	requestID, err := uuid.FromString(ctx.Param("requestId"))
	if err != nil {
		return err
	}

	if err := c.cs.AcceptFriendRequest(personnelID, requestID); err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusOK, nil)
}

// GetFriendships retrieves the user's friendships.
//
//	@Summary		Get friendships
//	@Description	Retrieve a list of friendships for the authenticated user
//	@Tags			Friends
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer"
//	@Success		200				{object}	domain.BaseResponse{data=[]domain.Friendship}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/friends [get]
func (c ChatController) GetFriendships(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}

	friendships, err := c.cs.GetFriendships(personnelID)
	if err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusOK, friendships)
}

// CreateConversation creates a new conversation.
//
//	@Summary		Create a conversation
//	@Description	Create a new conversation for the authenticated user
//	@Tags			Conversation
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string							true	"Bearer"
//	@Param			body			body		domain.CreateConversationInput	true	"Conversation creation input"
//	@Success		201				{object}	domain.BaseResponse{data=domain.Conversation}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/conversations [post]
func (c ChatController) CreateConversation(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}

	var input domain.CreateConversationInput
	if err := transport.DecodeAndValidateRequestBody(ctx, &input); err != nil {
		return err
	}

	result, err := c.cs.CreateConversation(personnelID, input)
	if err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusCreated, result)
}

// GetConversation retrieves a specific conversation.
//
//	@Summary		Get a conversation
//	@Description	Retrieve a specific conversation by ID
//	@Tags			Conversation
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer"
//	@Param			conversationId	path		string	true	"Conversation ID"
//	@Success		200				{object}	domain.BaseResponse{data=domain.Conversation}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/conversations/{conversationId} [get]
func (c ChatController) GetConversation(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}
	conversationID, err := uuid.FromString(ctx.Param("conversationId"))
	if err != nil {
		return err
	}

	conversation, err := c.cs.GetConversation(personnelID, conversationID)
	if err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusOK, conversation)
}

// GetUserConversations retrieves all conversations for a user.
//
//	@Summary		Get user conversations
//	@Description	Retrieve all conversations for the authenticated user
//	@Tags			Conversation
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer"
//	@Success		200				{object}	domain.BaseResponse{data=[]domain.Conversation}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/conversations [get]
func (c ChatController) GetUserConversations(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}

	conversations, err := c.cs.GetUserConversations(personnelID)
	if err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusOK, conversations)
}

// SendMessage sends a message to a conversation.
//
//	@Summary		Send a message
//	@Description	Send a message to a specific conversation
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string						true	"Bearer"
//	@Param			body			body		domain.CreateMessageInput	true	"Message creation input"
//	@Success		201				{object}	domain.BaseResponse{data=domain.Message}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/messages [post]
func (c ChatController) SendMessage(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}

	var input domain.CreateMessageInput
	if err := transport.DecodeAndValidateRequestBody(ctx, &input); err != nil {
		return err
	}

	result, err := c.cs.SendMessage(personnelID, input)
	if err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusCreated, result)
}

// GetMessages retrieves messages from a conversation.
//
//	@Summary		Get messages
//	@Description	Retrieve messages from a specific conversation
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer"
//	@Param			conversationId	path		string	true	"Conversation ID"
//	@Success		200				{object}	domain.BaseResponse{data=[]domain.Message}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/messages/{conversationId} [get]
func (c ChatController) GetMessages(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}
	conversationID, err := uuid.FromString(ctx.Param("conversationId"))
	if err != nil {
		return err
	}
	opt := transport.DecodeQueryOptions(ctx)

	messages, total, err := c.cs.GetMessages(personnelID, conversationID, opt)
	if err != nil {
		return err
	}

	return transport.SendPaginationResponse(ctx, http.StatusOK, messages, total)
}

// UpdateMessageStatus updates the status of a message.
//
//	@Summary		Update message status
//	@Description	Update the status of a specific message
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string							true	"Bearer"
//	@Param			messageId		path		string							true	"Message ID"
//	@Param			body			body		domain.UpdateMessageStatusInput	true	"Message status update input"
//	@Success		200				{object}	domain.BaseResponse{data=nil}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/messages/{messageId} [put]
func (c ChatController) UpdateMessageStatus(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}
	messageID, err := uuid.FromString(ctx.Param("messageId"))
	if err != nil {
		return err
	}

	var input domain.UpdateMessageStatusInput
	if err := transport.DecodeAndValidateRequestBody(ctx, &input); err != nil {
		return err
	}

	if err := c.cs.UpdateMessageStatus(personnelID, messageID, input); err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusOK, nil)
}

// StartMessaging starts a new messaging session.
//
//	@Summary		Start messaging
//	@Description	Start a new messaging session
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			id	query		string	true	"Personnel ID"
//	@Success		200	{object}	domain.BaseResponse{data=nil}
//	@Failure		400	{object}	domain.InvalidRequestError
//	@Failure		401	{object}	domain.UnauthorizedError
//	@Failure		500	{object}	domain.SystemError
//	@Router			/ws [get]
func (c ChatController) StartMessaging(ctx echo.Context) error {
	personnelID := ctx.QueryParam("id")

	if personnelID == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "missing personnel id"})
	}
	return transport.ServeWebSocket(c.hub, ctx, personnelID)
}
