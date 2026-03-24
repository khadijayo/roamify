package wishlist

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	// Items
	CreateItem(item *WishlistItem) error
	FindItemByID(id uuid.UUID) (*WishlistItem, error)
	FindItemsByUser(userID uuid.UUID) ([]WishlistItem, error)
	UpdateItem(item *WishlistItem) error
	DeleteItem(id uuid.UUID) error

	// Collections
	CreateCollection(c *WishlistCollection) error
	FindCollectionByID(id uuid.UUID) (*WishlistCollection, error)
	FindCollectionsByUser(userID uuid.UUID) ([]WishlistCollection, error)
	UpdateCollection(c *WishlistCollection) error
	DeleteCollection(id uuid.UUID) error

	// Collection-Item mapping
	AddItemToCollection(link *WishlistCollectionItem) error
	RemoveItemFromCollection(collectionID, itemID uuid.UUID) error
	FindCollectionItems(collectionID uuid.UUID) ([]WishlistItem, error)
	LinkExists(collectionID, itemID uuid.UUID) bool
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateItem(item *WishlistItem) error {
	return r.db.Create(item).Error
}

func (r *repository) FindItemByID(id uuid.UUID) (*WishlistItem, error) {
	var item WishlistItem
	err := r.db.First(&item, "id = ?", id).Error
	return &item, err
}

func (r *repository) FindItemsByUser(userID uuid.UUID) ([]WishlistItem, error) {
	var items []WishlistItem
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *repository) UpdateItem(item *WishlistItem) error {
	return r.db.Save(item).Error
}

func (r *repository) DeleteItem(id uuid.UUID) error {
	return r.db.Delete(&WishlistItem{}, "id = ?", id).Error
}

func (r *repository) CreateCollection(c *WishlistCollection) error {
	return r.db.Create(c).Error
}

func (r *repository) FindCollectionByID(id uuid.UUID) (*WishlistCollection, error) {
	var col WishlistCollection
	err := r.db.Preload("Items").First(&col, "id = ?", id).Error
	return &col, err
}

func (r *repository) FindCollectionsByUser(userID uuid.UUID) ([]WishlistCollection, error) {
	var cols []WishlistCollection
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&cols).Error
	return cols, err
}

func (r *repository) UpdateCollection(c *WishlistCollection) error {
	return r.db.Save(c).Error
}

func (r *repository) DeleteCollection(id uuid.UUID) error {
	return r.db.Delete(&WishlistCollection{}, "id = ?", id).Error
}

func (r *repository) AddItemToCollection(link *WishlistCollectionItem) error {
	return r.db.Create(link).Error
}

func (r *repository) RemoveItemFromCollection(collectionID, itemID uuid.UUID) error {
	return r.db.Where("collection_id = ? AND wishlist_item_id = ?", collectionID, itemID).
		Delete(&WishlistCollectionItem{}).Error
}

func (r *repository) FindCollectionItems(collectionID uuid.UUID) ([]WishlistItem, error) {
	var items []WishlistItem
	err := r.db.
		Joins("JOIN wishlist_collection_items wci ON wci.wishlist_item_id = wishlist_items.id").
		Where("wci.collection_id = ?", collectionID).
		Find(&items).Error
	return items, err
}

func (r *repository) LinkExists(collectionID, itemID uuid.UUID) bool {
	var count int64
	r.db.Model(&WishlistCollectionItem{}).
		Where("collection_id = ? AND wishlist_item_id = ?", collectionID, itemID).
		Count(&count)
	return count > 0
}