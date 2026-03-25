package wishlist

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khadijayo/roamify/pkg/middleware"
	"github.com/khadijayo/roamify/pkg/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.CreateItem(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, "wishlist item saved", item)
}

func (h *Handler) GetItems(c *gin.Context) {
	userID := middleware.GetUserID(c)
	items, err := h.svc.GetItems(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "wishlist items fetched", items)
}

func (h *Handler) UpdateItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}
	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.UpdateItem(itemID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "wishlist item updated", item)
}

func (h *Handler) DeleteItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}
	if err := h.svc.DeleteItem(itemID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "wishlist item deleted", nil)
}

func (h *Handler) CreateCollection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	col, err := h.svc.CreateCollection(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, "collection created", col)
}

func (h *Handler) GetCollections(c *gin.Context) {
	userID := middleware.GetUserID(c)
	cols, err := h.svc.GetCollections(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "collections fetched", cols)
}

func (h *Handler) GetCollection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	colID, err := uuid.Parse(c.Param("collectionId"))
	if err != nil {
		response.BadRequest(c, "invalid collection id")
		return
	}
	col, err := h.svc.GetCollection(colID, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, "collection fetched", col)
}

func (h *Handler) UpdateCollection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	colID, err := uuid.Parse(c.Param("collectionId"))
	if err != nil {
		response.BadRequest(c, "invalid collection id")
		return
	}
	var req UpdateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	col, err := h.svc.UpdateCollection(colID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "collection updated", col)
}

func (h *Handler) DeleteCollection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	colID, err := uuid.Parse(c.Param("collectionId"))
	if err != nil {
		response.BadRequest(c, "invalid collection id")
		return
	}
	if err := h.svc.DeleteCollection(colID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "collection deleted", nil)
}

func (h *Handler) AddItemToCollection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	colID, err := uuid.Parse(c.Param("collectionId"))
	if err != nil {
		response.BadRequest(c, "invalid collection id")
		return
	}
	var req AddToCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.AddItemToCollection(colID, userID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "item added to collection", nil)
}

func (h *Handler) RemoveItemFromCollection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	colID, err := uuid.Parse(c.Param("collectionId"))
	if err != nil {
		response.BadRequest(c, "invalid collection id")
		return
	}
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}
	if err := h.svc.RemoveItemFromCollection(colID, itemID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "item removed from collection", nil)
}
