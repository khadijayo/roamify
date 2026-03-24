package wishlist

import (
	"errors"

	"github.com/google/uuid"
)

type Service interface {
	// Items
	CreateItem(userID uuid.UUID, req *CreateItemRequest) (*WishlistItem, error)
	GetItems(userID uuid.UUID) ([]WishlistItem, error)
	UpdateItem(itemID, userID uuid.UUID, req *UpdateItemRequest) (*WishlistItem, error)
	DeleteItem(itemID, userID uuid.UUID) error

	// Collections
	CreateCollection(userID uuid.UUID, req *CreateCollectionRequest) (*WishlistCollection, error)
	GetCollections(userID uuid.UUID) ([]WishlistCollection, error)
	GetCollection(collectionID, userID uuid.UUID) (*WishlistCollection, error)
	UpdateCollection(collectionID, userID uuid.UUID, req *UpdateCollectionRequest) (*WishlistCollection, error)
	DeleteCollection(collectionID, userID uuid.UUID) error

	// Collection-Item mapping
	AddItemToCollection(collectionID, userID uuid.UUID, req *AddToCollectionRequest) error
	RemoveItemFromCollection(collectionID, itemID, userID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ---- Items ----

func (s *service) CreateItem(userID uuid.UUID, req *CreateItemRequest) (*WishlistItem, error) {
	item := &WishlistItem{
		UserID:        userID,
		Name:          req.Name,
		Location:      req.Location,
		Category:      req.Category,
		Notes:         req.Notes,
		EstimatedCost: req.EstimatedCost,
	}
	return item, s.repo.CreateItem(item)
}

func (s *service) GetItems(userID uuid.UUID) ([]WishlistItem, error) {
	return s.repo.FindItemsByUser(userID)
}

func (s *service) UpdateItem(itemID, userID uuid.UUID, req *UpdateItemRequest) (*WishlistItem, error) {
	item, err := s.repo.FindItemByID(itemID)
	if err != nil {
		return nil, err
	}
	if item.UserID != userID {
		return nil, errors.New("not authorized")
	}
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Location != "" {
		item.Location = req.Location
	}
	if req.Category != "" {
		item.Category = req.Category
	}
	if req.Notes != nil {
		item.Notes = req.Notes
	}
	if req.EstimatedCost != nil {
		item.EstimatedCost = req.EstimatedCost
	}
	return item, s.repo.UpdateItem(item)
}

func (s *service) DeleteItem(itemID, userID uuid.UUID) error {
	item, err := s.repo.FindItemByID(itemID)
	if err != nil {
		return err
	}
	if item.UserID != userID {
		return errors.New("not authorized")
	}
	return s.repo.DeleteItem(itemID)
}

// ---- Collections ----

func (s *service) CreateCollection(userID uuid.UUID, req *CreateCollectionRequest) (*WishlistCollection, error) {
	col := &WishlistCollection{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Emoji:       req.Emoji,
		ColorToken:  req.ColorToken,
	}
	return col, s.repo.CreateCollection(col)
}

func (s *service) GetCollections(userID uuid.UUID) ([]WishlistCollection, error) {
	return s.repo.FindCollectionsByUser(userID)
}

func (s *service) GetCollection(collectionID, userID uuid.UUID) (*WishlistCollection, error) {
	col, err := s.repo.FindCollectionByID(collectionID)
	if err != nil {
		return nil, err
	}
	if col.UserID != userID {
		return nil, errors.New("not authorized")
	}
	return col, nil
}

func (s *service) UpdateCollection(collectionID, userID uuid.UUID, req *UpdateCollectionRequest) (*WishlistCollection, error) {
	col, err := s.repo.FindCollectionByID(collectionID)
	if err != nil {
		return nil, err
	}
	if col.UserID != userID {
		return nil, errors.New("not authorized")
	}
	if req.Name != "" {
		col.Name = req.Name
	}
	if req.Description != nil {
		col.Description = req.Description
	}
	if req.Emoji != "" {
		col.Emoji = req.Emoji
	}
	if req.ColorToken != "" {
		col.ColorToken = req.ColorToken
	}
	return col, s.repo.UpdateCollection(col)
}

func (s *service) DeleteCollection(collectionID, userID uuid.UUID) error {
	col, err := s.repo.FindCollectionByID(collectionID)
	if err != nil {
		return err
	}
	if col.UserID != userID {
		return errors.New("not authorized")
	}
	return s.repo.DeleteCollection(collectionID)
}

func (s *service) AddItemToCollection(collectionID, userID uuid.UUID, req *AddToCollectionRequest) error {
	col, err := s.repo.FindCollectionByID(collectionID)
	if err != nil {
		return err
	}
	if col.UserID != userID {
		return errors.New("not authorized")
	}
	if s.repo.LinkExists(collectionID, req.WishlistItemID) {
		return errors.New("item already in collection")
	}
	link := &WishlistCollectionItem{
		CollectionID:   collectionID,
		WishlistItemID: req.WishlistItemID,
	}
	return s.repo.AddItemToCollection(link)
}

func (s *service) RemoveItemFromCollection(collectionID, itemID, userID uuid.UUID) error {
	col, err := s.repo.FindCollectionByID(collectionID)
	if err != nil {
		return err
	}
	if col.UserID != userID {
		return errors.New("not authorized")
	}
	return s.repo.RemoveItemFromCollection(collectionID, itemID)
}