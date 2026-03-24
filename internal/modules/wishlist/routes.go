package wishlist

import (
	"github.com/gin-gonic/gin"
)


func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	w := r.Group("/wishlist", auth)
	{
		// Spots / Items
		w.POST("/items", h.CreateItem)
		w.GET("/items", h.GetItems)
		w.PATCH("/items/:itemId", h.UpdateItem)
		w.DELETE("/items/:itemId", h.DeleteItem)

		// Collections
		w.POST("/collections", h.CreateCollection)
		w.GET("/collections", h.GetCollections)
		w.GET("/collections/:collectionId", h.GetCollection)
		w.PATCH("/collections/:collectionId", h.UpdateCollection)
		w.DELETE("/collections/:collectionId", h.DeleteCollection)

		// Collection-Item mapping
		w.POST("/collections/:collectionId/items", h.AddItemToCollection)
		w.DELETE("/collections/:collectionId/items/:itemId", h.RemoveItemFromCollection)
	}
}