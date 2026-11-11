package controller

import (
	"net/http"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/http/transport"
	"github.com/Rizwank123/chatApp/internal/pkg/security"
)

type SearchController struct {
	ss domain.SearchService
}

func NewSearchController(ss domain.SearchService) SearchController {
	return SearchController{ss: ss}
}

func getViewerID(c echo.Context) uuid.UUID {
	claims := security.GetClaimsForContext(c)
	if claims == nil {
		return uuid.Nil
	}
	raw, ok := claims["personnel_id"].(string)
	if !ok || raw == "" {
		return uuid.Nil
	}
	id, _ := uuid.FromString(raw)
	return id
}

// UnifiedSearch performs multi-category search across users, posts, and hashtags.
func (c SearchController) UnifiedSearch(ctx echo.Context) error {
	query := ctx.QueryParam("q")
	limitStr := ctx.QueryParam("limit")
	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	viewerID := getViewerID(ctx)
	results, err := c.ss.UnifiedSearch(ctx.Request().Context(), viewerID, query, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return transport.SendResponse(ctx, http.StatusOK, results)
}

// SearchUsers searches users with similarity and prefix matching.
func (c SearchController) SearchUsers(ctx echo.Context) error {
	query := ctx.QueryParam("q")
	opt := transport.DecodeQueryOptions(ctx)
	viewerID := getViewerID(ctx)

	results, total, err := c.ss.SearchUsers(ctx.Request().Context(), viewerID, query, int(opt.Limit), int(opt.Offset))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return transport.SendPaginationResponse(ctx, http.StatusOK, results, total)
}

// SearchPosts searches posts via full-text vector and trigrams.
func (c SearchController) SearchPosts(ctx echo.Context) error {
	query := ctx.QueryParam("q")
	opt := transport.DecodeQueryOptions(ctx)
	viewerID := getViewerID(ctx)

	results, total, err := c.ss.SearchPosts(ctx.Request().Context(), viewerID, query, int(opt.Limit), int(opt.Offset))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return transport.SendPaginationResponse(ctx, http.StatusOK, results, total)
}

// SearchHashtags searches hashtags by name.
func (c SearchController) SearchHashtags(ctx echo.Context) error {
	query := ctx.QueryParam("q")
	opt := transport.DecodeQueryOptions(ctx)

	results, total, err := c.ss.SearchHashtags(ctx.Request().Context(), query, int(opt.Limit), int(opt.Offset))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return transport.SendPaginationResponse(ctx, http.StatusOK, results, total)
}

// GetHashtagPosts retrieves all public posts with a specific hashtag.
func (c SearchController) GetHashtagPosts(ctx echo.Context) error {
	tag := ctx.Param("tag")
	opt := transport.DecodeQueryOptions(ctx)
	viewerID := getViewerID(ctx)

	results, total, err := c.ss.GetHashtagPosts(ctx.Request().Context(), viewerID, tag, int(opt.Limit), int(opt.Offset))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return transport.SendPaginationResponse(ctx, http.StatusOK, results, total)
}

// GetHashtagDetails retrieves hashtag stats.
func (c SearchController) GetHashtagDetails(ctx echo.Context) error {
	tag := ctx.Param("tag")
	details, err := c.ss.GetHashtagDetails(ctx.Request().Context(), tag)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "hashtag not found")
	}

	return transport.SendResponse(ctx, http.StatusOK, details)
}
