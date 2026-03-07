package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"roamify/internal/models"
)

type TipHandler struct {
	DB *gorm.DB
}

func NewTipHandler(db *gorm.DB) *TipHandler {
	return &TipHandler{DB: db}
}

// GET /api/v1/tips
func (h *TipHandler) List(c *gin.Context) {
	var tips []models.TravelTip

	query := h.DB.Where("is_published = true").Order("created_at desc")

	if wid := c.Query("wilaya_id"); wid != "" {
		query = query.Where("wilaya_id = ?", wid)
	}

	if q := c.Query("q"); q != "" {
		query = query.Where("LOWER(title) LIKE LOWER(?)", "%"+q+"%")
	}

	page, limit := getPagination(c)
	var total int64
	query.Model(&models.TravelTip{}).Count(&total)
	query = query.Offset((page - 1) * limit).Limit(limit)

	if err := query.Find(&tips).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tips"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tips,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GET /api/v1/tips/:slug
func (h *TipHandler) GetBySlug(c *gin.Context) {
	var tip models.TravelTip

	if err := h.DB.
		Where("slug = ? AND is_published = true", c.Param("slug")).
		First(&tip).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tip not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tip})
}