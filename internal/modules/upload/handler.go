package upload

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/internal/services"
	"github.com/khadijayo/roamify/pkg/response"
)

type Handler struct {
	cloud *services.CloudinaryService
}

func NewHandler(cloud *services.CloudinaryService) *Handler {
	return &Handler{cloud: cloud}
}

func (h *Handler) Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			response.BadRequest(c, "image file is required")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.InternalError(c, "failed to read uploaded file")
		return
	}
	defer file.Close()

	result, err := h.cloud.Upload(c.Request.Context(), file, fileHeader.Filename)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, "image uploaded", gin.H{
		"image_url": result.SecureURL,
		"public_id": result.PublicID,
	})
}
