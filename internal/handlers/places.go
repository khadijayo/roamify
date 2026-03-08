package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"roamify/internal/models"
)

type PlaceHandler struct {
	DB *gorm.DB
}

func NewPlaceHandler(db *gorm.DB) *PlaceHandler {
	return &PlaceHandler{DB: db}
}

// GET /api/v1/places
func (h *PlaceHandler) List(c *gin.Context) {
	var places []models.Place

	query := h.DB.
		Preload("Images").
		Preload("Wilaya").
		Preload("Activities").
		Preload("BestSeasons").
		Where("is_active = true")

	if cat := c.Query("category"); cat != "" {
		query = query.Where("category = ?", cat)
	}

	if wid := c.Query("wilaya_id"); wid != "" {
		query = query.Where("wilaya_id = ?", wid)
	}

	if rating := c.Query("rating"); rating != "" {
		query = query.Where("average_rating >= ?", rating)
	}

	if activity := c.Query("activity"); activity != "" {
		query = query.
			Joins("JOIN place_activities pa ON pa.place_id = places.id").
			Where("pa.activity = ?", activity)
	}

	if season := c.Query("season"); season != "" {
		query = query.
			Joins("JOIN place_seasons ps ON ps.place_id = places.id").
			Where("ps.season = ?", season)
	}

	if destType := c.Query("type"); destType != "" {
		query = query.
			Joins("JOIN destination_details dd ON dd.place_id = places.id").
			Where("dd.type = ?", destType)
	}

	if q := c.Query("q"); q != "" {
		query = query.Where("LOWER(places.name) LIKE LOWER(?)", "%"+q+"%")
	}

	// Pagination
	page, limit := getPagination(c)
	var total int64
	query.Model(&models.Place{}).Count(&total)
	query = query.Offset((page - 1) * limit).Limit(limit)

	if err := query.Find(&places).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch places"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  places,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GET /api/v1/places/featured
func (h *PlaceHandler) Featured(c *gin.Context) {
	var places []models.Place

	if err := h.DB.
		Preload("Images").
		Preload("Wilaya").
		Preload("Activities").
		Preload("BestSeasons").
		Where("is_featured = true AND is_active = true").
		Limit(6).
		Find(&places).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch featured places"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": places})
}

// GET /api/v1/places/:slug
func (h *PlaceHandler) GetBySlug(c *gin.Context) {
	var place models.Place

	err := h.DB.
		Preload("Images").
		Preload("Wilaya").
		Preload("Activities").
		Preload("BestSeasons").
		Preload("DestinationDetail").
		Preload("HotelDetail").
		Preload("RestaurantDetail").
		Preload("AgencyDetail").
		Where("slug = ? AND is_active = true", c.Param("slug")).
		First(&place).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "place not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": place})
}

// ── Pagination helper ──

func getPagination(c *gin.Context) (int, int) {
	page, limit := 1, 12

	if p := c.Query("page"); p != "" {
		if n, err := parseIntParam(p); err == nil && n > 0 {
			page = n
		}
	}
	if l := c.Query("limit"); l != "" {
		if n, err := parseIntParam(l); err == nil && n > 0 {
			limit = n
			if limit > 50 {
				limit = 50
			}
		}
	}
	return page, limit
}

func parseIntParam(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}