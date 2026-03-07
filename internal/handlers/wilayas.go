package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"roamify/internal/models"
)

type WilayaHandler struct {
	DB *gorm.DB
}

func NewWilayaHandler(db *gorm.DB) *WilayaHandler {
	return &WilayaHandler{DB: db}
}

// GET /api/v1/wilayas
func (h *WilayaHandler) List(c *gin.Context) {
	var wilayas []models.Wilaya

	if err := h.DB.Order("code asc").Find(&wilayas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch wilayas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wilayas})
}

// GET /api/v1/wilayas/:id
func (h *WilayaHandler) GetByID(c *gin.Context) {
	var wilaya models.Wilaya

	if err := h.DB.
		Preload("Places", "is_active = true").
		First(&wilaya, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wilaya not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wilaya})
}