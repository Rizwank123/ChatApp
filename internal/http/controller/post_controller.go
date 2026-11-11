package controller

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/http/transport"
	"github.com/Rizwank123/chatApp/internal/pkg/security"
)

type PostController struct {
	ps domain.PostService
}

func NewPostController(ps domain.PostService) PostController {
	return PostController{ps: ps}
}

// CreatePost creates a new post.
//
//	@Summary		Create new post
//	@Description	Create a new post for the authenticated personnel
//	@Tags			Post
//	@ID				createPost
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string					true	"Bearer "
//	@Param			body			body		domain.CreatePostInput	true	"Post creation input"
//	@Success		201				{object}	domain.BaseResponse{data=domain.Post}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/posts [post]
func (c PostController) CreatePost(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)

	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}

	var input domain.CreatePostInput
	if err := transport.DecodeAndValidateRequestBody(ctx, &input); err != nil {
		return err
	}

	post, err := c.ps.CreatePost(input, personnelID)
	if err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusCreated, post)
}

// GetPost retrieves a post by ID.
//
//	@Summary		Get post by ID
//	@Description	Get post details based on the provided ID
//	@Tags			Post
//	@ID				getPost
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer "
//	@Param			id				path		string	true	"Post ID"
//	@Success		200				{object}	domain.BaseResponse{data=domain.Post}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		404				{object}	domain.NotFoundError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/posts/{id} [get]
func (c PostController) GetPost(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}

	claims := security.GetClaimsForContext(ctx)
	var viewerID uuid.UUID
	if claims != nil {
		if rawID, ok := claims["personnel_id"].(string); ok && rawID != "" {
			viewerID, _ = uuid.FromString(rawID)
		}
	}

	result, err := c.ps.GetPost(id, viewerID)
	if err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusOK, result)
}

// GetUserPosts retrieves posts by user ID with pagination.
//
//	@Summary		Get user posts
//	@Description	Get posts of a specific user with pagination
//	@Tags			Post
//	@ID				getUserPosts
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer "
//	@Param			id				path		string	true	"User ID"
//	@Param			page			query		int		false	"Page number"
//	@Param			limit			query		int		false	"Items per page"
//	@Success		200				{object}	domain.BaseResponse{data=[]domain.Post}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/users/{id}/posts [get]
func (c PostController) GetUserPosts(ctx echo.Context) error {
	userID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}

	claims := security.GetClaimsForContext(ctx)
	var viewerID uuid.UUID
	if claims != nil {
		if rawID, ok := claims["personnel_id"].(string); ok && rawID != "" {
			viewerID, _ = uuid.FromString(rawID)
		}
	}

	opt := transport.DecodeQueryOptions(ctx)

	result, total, err := c.ps.GetUserPosts(userID, opt, viewerID)
	if err != nil {
		return err
	}

	return transport.SendPaginationResponse(ctx, http.StatusOK, result, total)
}

// GetPublicPosts retrieves explore/discovery posts.
func (c PostController) GetPublicPosts(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	var viewerID uuid.UUID
	if claims != nil {
		if rawID, ok := claims["personnel_id"].(string); ok && rawID != "" {
			viewerID, _ = uuid.FromString(rawID)
		}
	}

	opt := transport.DecodeQueryOptions(ctx)
	result, total, err := c.ps.GetPublicPosts(opt, viewerID)
	if err != nil {
		return err
	}

	return transport.SendPaginationResponse(ctx, http.StatusOK, result, total)
}

// RecordView increments a post view count.
func (c PostController) RecordView(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}
	_ = c.ps.RecordView(id)
	return transport.SendResponse(ctx, http.StatusOK, map[string]bool{"success": true})
}

// RecordShare increments a post share count.
func (c PostController) RecordShare(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}
	_ = c.ps.RecordShare(id)
	return transport.SendResponse(ctx, http.StatusOK, map[string]bool{"success": true})
}

// GetFeedPosts retrieves feed posts for authenticated user.
//
//	@Summary		Get feed posts
//	@Description	Get feed posts for the authenticated user with pagination
//	@Tags			Post
//	@ID				getFeedPosts
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer "
//	@Param			page			query		int		false	"Page number"
//	@Param			limit			query		int		false	"Items per page"
//	@Success		200				{object}	domain.BaseResponse{data=[]domain.Post}
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/posts/feed [get]
func (c PostController) GetFeedPosts(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)

	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}

	opt := transport.DecodeQueryOptions(ctx)

	result, total, err := c.ps.GetFeedPosts(personnelID, opt)
	if err != nil {
		return err
	}

	return transport.SendPaginationResponse(ctx, http.StatusOK, result, total)
}

// LikePost likes a post.
//
//	@Summary		Like post
//	@Description	Like a post for the authenticated user
//	@Tags			Post
//	@ID				likePost
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer "
//	@Param			id				path		string	true	"Post ID"
//	@Success		204				{object}	nil
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/posts/{id}/like [post]
func (c PostController) LikePost(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)

	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}

	postID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}

	if err := c.ps.LikePost(personnelID, postID); err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusNoContent, nil)
}

// UnlikePost unlikes a post.
//
//	@Summary		Unlike post
//	@Description	Unlike a post for the authenticated user
//	@Tags			Post
//	@ID				unlikePost
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer "
//	@Param			id				path		string	true	"Post ID"
//	@Success		204				{object}	nil
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/posts/{id}/unlike [post]
func (c PostController) UnlikePost(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)

	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}

	postID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}

	if err := c.ps.UnlikePost(personnelID, postID); err != nil {
		return err
	}

	return transport.SendResponse(ctx, http.StatusNoContent, nil)
}

func (c PostController) UpdatePost(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}
	postID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}

	var input domain.UpdatePostInput
	if err := transport.DecodeAndValidateRequestBody(ctx, &input); err != nil {
		return err
	}
	if err := c.ps.UpdatePost(postID, input, personnelID); err != nil {
		return err
	}
	return transport.SendResponse(ctx, http.StatusOK, nil)
}

func (c PostController) DeletePost(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}
	postID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}
	if err := c.ps.DeletePost(postID, personnelID); err != nil {
		return err
	}
	return transport.SendResponse(ctx, http.StatusNoContent, nil)
}

func (c PostController) FollowUser(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}
	followingID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}
	if err := c.ps.FollowUser(personnelID, followingID); err != nil {
		return err
	}
	return transport.SendResponse(ctx, http.StatusNoContent, nil)
}

func (c PostController) UnfollowUser(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}
	followingID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}
	if err := c.ps.UnfollowUser(personnelID, followingID); err != nil {
		return err
	}
	return transport.SendResponse(ctx, http.StatusNoContent, nil)
}

func (c PostController) CreateComment(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}
	postID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}

	var input domain.CreateCommentInput
	if err := transport.DecodeAndValidateRequestBody(ctx, &input); err != nil {
		return err
	}
	if err := c.ps.CreateComment(postID, input.Content, personnelID, input.ParentID); err != nil {
		return err
	}
	return transport.SendResponse(ctx, http.StatusCreated, nil)
}

func (c PostController) DeleteComment(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	personnelID, err := uuid.FromString(claims["personnel_id"].(string))
	if err != nil {
		return err
	}
	commentID, err := uuid.FromString(ctx.Param("commentId"))
	if err != nil {
		return err
	}
	if err := c.ps.DeleteComment(commentID, personnelID); err != nil {
		return err
	}
	return transport.SendResponse(ctx, http.StatusNoContent, nil)
}

func (c PostController) GetPostComments(ctx echo.Context) error {
	postID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}
	opt := transport.DecodeQueryOptions(ctx)
	result, total, err := c.ps.GetPostComments(postID, opt)
	if err != nil {
		return err
	}
	return transport.SendPaginationResponse(ctx, http.StatusOK, result, total)
}
