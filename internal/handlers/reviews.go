package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"roamify/internal/middleware"
	"roamify/internal/models"
)

type ReviewHandler struct {
	DB *gorm.DB
}

func NewReviewHandler(db *gorm.DB) *ReviewHandler {
	return &ReviewHandler{DB: db}
}

type CreateReviewRequest struct {
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
	Title  string `json:"title"`
	Body   string `json:"body"   binding:"required,min=10"`
}

// GET /api/v1/places/:slug/reviews
func (h *ReviewHandler) List(c *gin.Context) {
	var place models.Place
	if err := h.DB.Where("slug = ?", c.Param("slug")).First(&place).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "place not found"})
		return
	}

	var reviews []models.Review
	h.DB.
		Preload("User").
		Where("place_id = ? AND is_visible = true", place.ID).
		Order("created_at desc").
		Find(&reviews)

	c.JSON(http.StatusOK, gin.H{"data": reviews, "count": len(reviews)})
}

// POST /api/v1/places/:slug/reviews  (protected)
func (h *ReviewHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var place models.Place
	if err := h.DB.Where("slug = ? AND is_active = true", c.Param("slug")).First(&place).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "place not found"})
		return
	}

	// Prevent duplicate review from same user
	var existing models.Review
	if err := h.DB.Where("place_id = ? AND user_id = ?", place.ID, userID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "you have already reviewed this place"})
		return
	}

	review := models.Review{
		ID:        uuid.New(),
		PlaceID:   place.ID,
		UserID:    userID,
		Rating:    req.Rating,
		Title:     req.Title,
		Body:      req.Body,
		IsVisible: true,
	}

	if err := h.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review"})
		return
	}

	h.updatePlaceRating(place.ID)

	c.JSON(http.StatusCreated, gin.H{"data": review})
}

// updatePlaceRating recalculates average_rating and review_count on the place
func (h *ReviewHandler) updatePlaceRating(placeID uuid.UUID) {
	type result struct {
		Avg   float32
		Count int
	}
	var res result
	h.DB.Model(&models.Review{}).
		Select("AVG(rating) as avg, COUNT(*) as count").
		Where("place_id = ? AND is_visible = true", placeID).
		Scan(&res)

	h.DB.Model(&models.Place{}).
		Where("id = ?", placeID).
		Updates(map[string]interface{}{
			"average_rating": res.Avg,
			"review_count":   res.Count,
		})
}