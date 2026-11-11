package controller

import (
	"net/http"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/pkg/security"
	"github.com/Rizwank123/chatApp/internal/service"
)

// SocialController handles HTTP requests for feed, social graph, blocking, bookmarks, and notifications.
type SocialController struct {
	ss service.SocialService
}

// NewSocialController constructs a new SocialController.
func NewSocialController(ss service.SocialService) SocialController {
	return SocialController{ss: ss}
}

func getPersonnelID(c echo.Context) (uuid.UUID, error) {
	claims := security.GetClaimsForContext(c)
	rawID, ok := claims["personnel_id"].(string)
	if !ok || rawID == "" {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid personnel_id in token")
	}
	return uuid.FromString(rawID)
}

// GetFeed retrieves the user's home feed generated via the hybrid fanout engine.
func (sc SocialController) GetFeed(c echo.Context) error {
	personnelID, err := getPersonnelID(c)
	if err != nil {
		return err
	}

	cursor := c.QueryParam("cursor")
	limitStr := c.QueryParam("limit")
	var limit int64 = 20
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil && l > 0 {
			limit = l
		}
	}

	page, err := sc.ss.GetFeed(c.Request().Context(), personnelID, cursor, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	posts := make([]domain.Post, len(page.Items))
	for i, it := range page.Items {
		posts[i] = it.Post
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data":        posts,
		"items":       page.Items,
		"next_cursor": page.NextCursor,
		"has_more":    page.HasMore,
		"total":       len(posts),
	})
}

// FollowUser follows a user.
func (sc SocialController) FollowUser(c echo.Context) error {
	personnelID, err := getPersonnelID(c)
	if err != nil {
		return err
	}

	targetID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid target user id")
	}

	status, err := sc.ss.FollowUser(c.Request().Context(), personnelID, targetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, domain.BaseResponse{
		Data: map[string]string{
			"message":             "User follow updated",
			"relationship_status": status,
		},
	})
}

// UnfollowUser unfollows a user.
func (sc SocialController) UnfollowUser(c echo.Context) error {
	personnelID, err := getPersonnelID(c)
	if err != nil {
		return err
	}

	targetID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid target user id")
	}

	if err := sc.ss.UnfollowUser(c.Request().Context(), personnelID, targetID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, domain.BaseResponse{
		Data: map[string]string{"message": "Unfollowed successfully"},
	})
}

// BlockUser blocks a user.
func (sc SocialController) BlockUser(c echo.Context) error {
	personnelID, err := getPersonnelID(c)
	if err != nil {
		return err
	}

	targetID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid target user id")
	}

	if err := sc.ss.BlockUser(c.Request().Context(), personnelID, targetID, "user blocked"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, domain.BaseResponse{
		Data: map[string]string{"message": "User blocked successfully"},
	})
}

// UnblockUser removes a block.
func (sc SocialController) UnblockUser(c echo.Context) error {
	personnelID, err := getPersonnelID(c)
	if err != nil {
		return err
	}

	targetID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid target user id")
	}

	if err := sc.ss.UnblockUser(c.Request().Context(), personnelID, targetID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, domain.BaseResponse{
		Data: map[string]string{"message": "User unblocked successfully"},
	})
}

// SavePost bookmarks a post.
func (sc SocialController) SavePost(c echo.Context) error {
	personnelID, err := getPersonnelID(c)
	if err != nil {
		return err
	}

	postID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid post id")
	}

	if err := sc.ss.SavePost(c.Request().Context(), personnelID, postID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, domain.BaseResponse{
		Data: map[string]string{"message": "Post saved"},
	})
}

// UnsavePost removes a post bookmark.
func (sc SocialController) UnsavePost(c echo.Context) error {
	personnelID, err := getPersonnelID(c)
	if err != nil {
		return err
	}

	postID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid post id")
	}

	if err := sc.ss.UnsavePost(c.Request().Context(), personnelID, postID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, domain.BaseResponse{
		Data: map[string]string{"message": "Post removed from saved"},
	})
}

// GetNotifications returns user notifications.
func (sc SocialController) GetNotifications(c echo.Context) error {
	personnelID, err := getPersonnelID(c)
	if err != nil {
		return err
	}

	limit := 30
	offset := 0
	list, total, err := sc.ss.GetNotifications(c.Request().Context(), personnelID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, domain.PaginationResponse{
		Data:  list,
		Total: total,
		Size:  int64(limit),
		Page:  int64(offset/limit + 1),
	})
}

// MarkNotificationRead marks a notification as read.
func (sc SocialController) MarkNotificationRead(c echo.Context) error {
	personnelID, err := getPersonnelID(c)
	if err != nil {
		return err
	}

	notifID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid notification id")
	}

	if err := sc.ss.MarkNotificationRead(c.Request().Context(), notifID, personnelID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, domain.BaseResponse{
		Data: map[string]string{"message": "Notification marked as read"},
	})
}

// GetUserCounters returns follower, following, and post counts for a user.
func (sc SocialController) GetUserCounters(c echo.Context) error {
	idParam := c.Param("id")
	var targetID uuid.UUID
	var err error
	if idParam != "" && idParam != "me" {
		targetID, err = uuid.FromString(idParam)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
		}
	} else {
		targetID, err = getPersonnelID(c)
		if err != nil {
			return err
		}
	}

	counters, err := sc.ss.GetUserCounters(c.Request().Context(), targetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, domain.BaseResponse{Data: counters})
}

// GetRelationship returns follow and chat relationship between the authenticated user and target user.
func (sc SocialController) GetRelationship(c echo.Context) error {
	personnelID, err := getPersonnelID(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid target user id")
	}

	rel, err := sc.ss.GetRelationship(c.Request().Context(), personnelID, targetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, domain.BaseResponse{Data: rel})
}

// CanChat returns whether two users are authorized to chat.
func (sc SocialController) CanChat(c echo.Context) error {
	personnelID, err := getPersonnelID(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid target user id")
	}

	rel, err := sc.ss.GetRelationship(c.Request().Context(), personnelID, targetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var reason string
	if !rel.CanChat {
		if !rel.IsFollowing && !rel.FollowsYou {
			reason = "Messaging requires either user to follow the other"
		} else {
			reason = "Messaging is unavailable due to an active block or account status"
		}
	}
	return c.JSON(http.StatusOK, domain.BaseResponse{
		Data: map[string]interface{}{
			"can_chat":     rel.CanChat,
			"is_mutual":    rel.IsMutual,
			"is_following": rel.IsFollowing,
			"follows_you":  rel.FollowsYou,
			"reason":       reason,
		},
	})
}

// GetFollowers returns the paginated list of users following the target user.
func (sc SocialController) GetFollowers(c echo.Context) error {
	currentID, err := getPersonnelID(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid target user id")
	}

	query := c.QueryParam("q")
	limit := 20
	if l, err := strconv.Atoi(c.QueryParam("limit")); err == nil && l > 0 {
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(c.QueryParam("offset")); err == nil && o >= 0 {
		offset = o
	}

	list, total, err := sc.ss.GetFollowersList(c.Request().Context(), currentID, targetID, query, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, domain.PaginationResponse{
		Data:  list,
		Total: total,
		Size:  int64(limit),
		Page:  int64(offset/limit + 1),
	})
}

// GetFollowing returns the paginated list of users that the target user is following.
func (sc SocialController) GetFollowing(c echo.Context) error {
	currentID, err := getPersonnelID(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid target user id")
	}

	query := c.QueryParam("q")
	limit := 20
	if l, err := strconv.Atoi(c.QueryParam("limit")); err == nil && l > 0 {
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(c.QueryParam("offset")); err == nil && o >= 0 {
		offset = o
	}

	list, total, err := sc.ss.GetFollowingList(c.Request().Context(), currentID, targetID, query, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, domain.PaginationResponse{
		Data:  list,
		Total: total,
		Size:  int64(limit),
		Page:  int64(offset/limit + 1),
	})
}

// SearchUsers searches users by username or display name with follow relationship statuses.
func (sc SocialController) SearchUsers(c echo.Context) error {
	currentID, err := getPersonnelID(c)
	if err != nil {
		return err
	}
	query := c.QueryParam("q")
	limit := 20
	if l, err := strconv.Atoi(c.QueryParam("limit")); err == nil && l > 0 {
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(c.QueryParam("offset")); err == nil && o >= 0 {
		offset = o
	}

	results, err := sc.ss.SearchUsers(c.Request().Context(), currentID, query, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, domain.BaseResponse{Data: results})
}
