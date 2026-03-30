package trips

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateTrip(trip *Trip) error
	FindTripByID(id uuid.UUID) (*Trip, error)
	FindTripsByOwner(ownerID uuid.UUID) ([]Trip, error)
	FindTripsByMember(userID uuid.UUID) ([]Trip, error)
	UpdateTrip(trip *Trip) error
	DeleteTrip(id uuid.UUID) error

	AddMember(member *TripMember) error
	FindMember(tripID, userID uuid.UUID) (*TripMember, error)
	FindMembersByTrip(tripID uuid.UUID) ([]TripMember, error)
	UpdateMember(member *TripMember) error
	RemoveMember(tripID, userID uuid.UUID) error

	CreateItineraryItem(item *TripItineraryItem) error
	FindItineraryByTrip(tripID uuid.UUID) ([]TripItineraryItem, error)
	FindItineraryItemByID(id uuid.UUID) (*TripItineraryItem, error)
	UpdateItineraryItem(item *TripItineraryItem) error
	DeleteItineraryItem(id uuid.UUID) error

	CreateExpense(expense *TripExpense) error
	FindExpensesByTrip(tripID uuid.UUID) ([]TripExpense, error)
	FindExpenseByID(id uuid.UUID) (*TripExpense, error)
	UpdateExpense(expense *TripExpense) error
	DeleteExpense(id uuid.UUID) error
	SumExpensesByTrip(tripID uuid.UUID) (float64, error)

	FindChatMessages(tripID uuid.UUID, limit int) ([]ChatMessage, error)
	CreateChatMessage(msg *ChatMessage) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateTrip(trip *Trip) error {
	return r.db.Create(trip).Error
}

func (r *repository) FindTripByID(id uuid.UUID) (*Trip, error) {
	var trip Trip
	err := r.db.
		Preload("Members").
		Preload("ItineraryItems").
		Preload("Expenses").
		First(&trip, "id = ?", id).Error
	return &trip, err
}

func (r *repository) FindTripsByOwner(ownerID uuid.UUID) ([]Trip, error) {
	var trips []Trip
	err := r.db.Where("owner_user_id = ?", ownerID).Order("created_at DESC").Find(&trips).Error
	return trips, err
}

func (r *repository) FindTripsByMember(userID uuid.UUID) ([]Trip, error) {
	var trips []Trip
	err := r.db.
		Joins("JOIN trip_members ON trip_members.trip_id = trips.id").
		Where("trip_members.user_id = ? AND trip_members.join_status = ?", userID, JoinStatusJoined).
		Order("trips.created_at DESC").
		Find(&trips).Error
	return trips, err
}

func (r *repository) UpdateTrip(trip *Trip) error {
	return r.db.Save(trip).Error
}

func (r *repository) DeleteTrip(id uuid.UUID) error {
	return r.db.Delete(&Trip{}, "id = ?", id).Error
}

func (r *repository) AddMember(member *TripMember) error {
	return r.db.Create(member).Error
}

func (r *repository) FindMember(tripID, userID uuid.UUID) (*TripMember, error) {
	var m TripMember
	err := r.db.Where("trip_id = ? AND user_id = ?", tripID, userID).First(&m).Error
	return &m, err
}

func (r *repository) FindMembersByTrip(tripID uuid.UUID) ([]TripMember, error) {
	var members []TripMember
	err := r.db.Where("trip_id = ?", tripID).Find(&members).Error
	return members, err
}

func (r *repository) UpdateMember(member *TripMember) error {
	return r.db.Save(member).Error
}

func (r *repository) RemoveMember(tripID, userID uuid.UUID) error {
	return r.db.Where("trip_id = ? AND user_id = ?", tripID, userID).Delete(&TripMember{}).Error
}

func (r *repository) CreateItineraryItem(item *TripItineraryItem) error {
	return r.db.Create(item).Error
}

func (r *repository) FindItineraryByTrip(tripID uuid.UUID) ([]TripItineraryItem, error) {
	var items []TripItineraryItem
	err := r.db.
		Where("trip_id = ?", tripID).
		Order("day_number ASC, sort_order ASC").
		Find(&items).Error
	return items, err
}

func (r *repository) FindItineraryItemByID(id uuid.UUID) (*TripItineraryItem, error) {
	var item TripItineraryItem
	err := r.db.First(&item, "id = ?", id).Error
	return &item, err
}

func (r *repository) UpdateItineraryItem(item *TripItineraryItem) error {
	return r.db.Save(item).Error
}

func (r *repository) DeleteItineraryItem(id uuid.UUID) error {
	return r.db.Delete(&TripItineraryItem{}, "id = ?", id).Error
}

func (r *repository) CreateExpense(expense *TripExpense) error {
	return r.db.Create(expense).Error
}

func (r *repository) FindExpensesByTrip(tripID uuid.UUID) ([]TripExpense, error) {
	var expenses []TripExpense
	err := r.db.Where("trip_id = ?", tripID).Order("expense_date DESC").Find(&expenses).Error
	return expenses, err
}

func (r *repository) FindExpenseByID(id uuid.UUID) (*TripExpense, error) {
	var expense TripExpense
	err := r.db.First(&expense, "id = ?", id).Error
	return &expense, err
}

func (r *repository) UpdateExpense(expense *TripExpense) error {
	return r.db.Save(expense).Error
}

func (r *repository) DeleteExpense(id uuid.UUID) error {
	return r.db.Delete(&TripExpense{}, "id = ?", id).Error
}

func (r *repository) SumExpensesByTrip(tripID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.Model(&TripExpense{}).
		Where("trip_id = ?", tripID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *repository) FindChatMessages(tripID uuid.UUID, limit int) ([]ChatMessage, error) {
	var msgs []ChatMessage
	err := r.db.
		Where("trip_id = ?", tripID).
		Order("created_at ASC").
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

func (r *repository) CreateChatMessage(msg *ChatMessage) error {
	return r.db.Create(msg).Error
}
