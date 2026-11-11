package controller

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/http/transport"
	"github.com/Rizwank123/chatApp/internal/pkg/security"
)

type MediaController struct {
	ms domain.MediaService
}

func NewMediaController(ms domain.MediaService) MediaController {
	return MediaController{ms: ms}
}

// InitiateUpload generates a presigned MinIO upload session.
func (c MediaController) InitiateUpload(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	rawID, ok := claims["personnel_id"].(string)
	if !ok || rawID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing personnel_id")
	}
	personnelID, err := uuid.FromString(rawID)
	if err != nil {
		return err
	}

	var input domain.InitiateUploadInput
	if err := transport.DecodeAndValidateRequestBody(ctx, &input); err != nil {
		return err
	}

	out, err := c.ms.InitiateUpload(ctx.Request().Context(), personnelID, input)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return transport.SendResponse(ctx, http.StatusCreated, out)
}

// CompleteUpload verifies the uploaded MinIO object and confirms ready state.
func (c MediaController) CompleteUpload(ctx echo.Context) error {
	claims := security.GetClaimsForContext(ctx)
	rawID, ok := claims["personnel_id"].(string)
	if !ok || rawID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing personnel_id")
	}
	personnelID, err := uuid.FromString(rawID)
	if err != nil {
		return err
	}

	uploadID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid upload id")
	}

	upload, err := c.ms.CompleteUpload(ctx.Request().Context(), personnelID, uploadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return transport.SendResponse(ctx, http.StatusOK, upload)
}

// GetMedia retrieves metadata and URLs for a post media object.
func (c MediaController) GetMedia(ctx echo.Context) error {
	mediaID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid media id")
	}

	media, err := c.ms.GetMedia(ctx.Request().Context(), mediaID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "media not found")
	}

	return transport.SendResponse(ctx, http.StatusOK, media)
}
