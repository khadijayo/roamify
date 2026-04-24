package reports

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khadijayo/roamify/pkg/middleware"
	"github.com/khadijayo/roamify/pkg/response"
	"gorm.io/gorm"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	report, err := h.svc.Create(c.Request.Context(), middleware.GetUserID(c), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTarget):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Created(c, "report created", report)
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	items, meta, err := h.svc.List(c.Request.Context(), c.Query("status"), page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OKPaginated(c, "reports fetched", items, meta)
}

func (h *Handler) Resolve(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	item, err := h.svc.Resolve(c.Request.Context(), reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "report not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "report resolved", item)
}
