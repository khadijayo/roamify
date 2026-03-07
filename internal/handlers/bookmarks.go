package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"roamify/internal/middleware"
	"roamify/internal/models"
)

type BookmarkHandler struct {
	DB *gorm.DB
}

func NewBookmarkHandler(db *gorm.DB) *BookmarkHandler {
	return &BookmarkHandler{DB: db}
}

// GET /api/v1/user/bookmarks  (protected)
func (h *BookmarkHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var bookmarks []models.Bookmark
	h.DB.
		Preload("Place.Images").
		Preload("Place.Wilaya").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&bookmarks)

	c.JSON(http.StatusOK, gin.H{"data": bookmarks, "count": len(bookmarks)})
}

// POST /api/v1/user/bookmarks/:placeID  (protected)
func (h *BookmarkHandler) Add(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	placeID, err := uuid.Parse(c.Param("placeID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid place id"})
		return
	}

	var place models.Place
	if err := h.DB.Where("id = ? AND is_active = true", placeID).First(&place).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "place not found"})
		return
	}

	var existing models.Bookmark
	if err := h.DB.Where("user_id = ? AND place_id = ?", userID, placeID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already bookmarked"})
		return
	}

	bookmark := models.Bookmark{
		UserID:  userID,
		PlaceID: placeID,
	}

	if err := h.DB.Create(&bookmark).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add bookmark"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "bookmarked successfully"})
}

// DELETE /api/v1/user/bookmarks/:placeID  (protected)
func (h *BookmarkHandler) Remove(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	placeID, err := uuid.Parse(c.Param("placeID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid place id"})
		return
	}

	result := h.DB.Where("user_id = ? AND place_id = ?", userID, placeID).Delete(&models.Bookmark{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "bookmark not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "bookmark removed"})
}