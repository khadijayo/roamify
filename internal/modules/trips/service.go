package trips

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Service interface {
	CreateTrip(ownerID uuid.UUID, req *CreateTripRequest) (*Trip, error)
	GetTrip(tripID, requesterID uuid.UUID) (*Trip, error)
	GetMyTrips(userID uuid.UUID) ([]Trip, error)
	UpdateTrip(tripID, requesterID uuid.UUID, req *UpdateTripRequest) (*Trip, error)
	DeleteTrip(tripID, requesterID uuid.UUID) error

	InviteMember(tripID, requesterID uuid.UUID, req *InviteMemberRequest) (*TripMember, error)
	UpdateMemberStatus(tripID, userID uuid.UUID, req *UpdateMemberStatusRequest) (*TripMember, error)
	RemoveMember(tripID, requesterID, targetUserID uuid.UUID) error
	GetMembers(tripID uuid.UUID) ([]TripMember, error)

	AddItineraryItem(tripID, userID uuid.UUID, req *CreateItineraryItemRequest) (*TripItineraryItem, error)
	GenerateAndSaveAIItinerary(tripID, userID uuid.UUID, req *GenerateAIItineraryRequest) ([]TripItineraryItem, error)
	PlanAndCreateTripWithAI(userID uuid.UUID, req *PlanAndCreateTripRequest) (*PlanAndCreateTripResponse, error)
	GetItinerary(tripID uuid.UUID) ([]TripItineraryItem, error)
	UpdateItineraryItem(itemID, requesterID uuid.UUID, req *UpdateItineraryItemRequest) (*TripItineraryItem, error)
	DeleteItineraryItem(itemID, requesterID uuid.UUID) error

	AddExpense(tripID, userID uuid.UUID, req *CreateExpenseRequest) (*TripExpense, error)
	GetExpenses(tripID uuid.UUID) ([]TripExpense, error)
	UpdateExpense(expenseID, requesterID uuid.UUID, req *UpdateExpenseRequest) (*TripExpense, error)
	DeleteExpense(expenseID, requesterID uuid.UUID) error

	GetChatHistory(tripID uuid.UUID, limit int) ([]ChatMessage, error)
	SendChatMessage(tripID, userID uuid.UUID, message string) (*ChatMessage, error)
	GetTripMapPins(tripID uuid.UUID) ([]MapPin, error)
}

type service struct {
	repo       Repository
	grokKey    string
	httpClient *http.Client
}

func NewService(repo Repository, grokKey string) Service {
	return &service{
		repo:       repo,
		grokKey:    grokKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *service) isMember(tripID, userID uuid.UUID) bool {
	trip, tripErr := s.repo.FindTripByID(tripID)
	if tripErr == nil && trip.OwnerUserID == userID {
		return true
	}

	m, err := s.repo.FindMember(tripID, userID)
	if err == nil {
		return m.JoinStatus == JoinStatusJoined
	}

	return false
}

func (s *service) isOwnerOrAdmin(tripID, userID uuid.UUID) bool {
	trip, tripErr := s.repo.FindTripByID(tripID)
	if tripErr == nil && trip.OwnerUserID == userID {
		return true
	}

	m, err := s.repo.FindMember(tripID, userID)
	if err == nil {
		return m.Role == RoleOwner || m.Role == RoleAdmin
	}

	return false
}

func (s *service) CreateTrip(ownerID uuid.UUID, req *CreateTripRequest) (*Trip, error) {
	travelers := req.TravelersPlanned
	if travelers < 1 {
		travelers = 1
	}
	trip := &Trip{
		OwnerUserID:      ownerID,
		Title:            req.Title,
		Destination:      req.Destination,
		TripType:         req.TripType,
		VibeTags:         pq.StringArray(req.VibeTags),
		TravelersPlanned: travelers,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		Budget:           req.Budget,
		CoverImageURL:    req.CoverImageURL,
		Notes:            req.Notes,
		Status:           TripStatusPlanning,
	}
	if err := s.repo.CreateTrip(trip); err != nil {
		return nil, err
	}

	now := time.Now()
	member := &TripMember{
		TripID:     trip.ID,
		UserID:     ownerID,
		Role:       RoleOwner,
		JoinStatus: JoinStatusJoined,
		JoinedAt:   &now,
	}
	_ = s.repo.AddMember(member)
	return trip, nil
}

func (s *service) GetTrip(tripID, requesterID uuid.UUID) (*Trip, error) {
	trip, err := s.repo.FindTripByID(tripID)
	if err != nil {
		return nil, err
	}

	if trip.OwnerUserID != requesterID && !s.isMember(tripID, requesterID) {
		return nil, errors.New("access denied")
	}
	return trip, nil
}

func (s *service) GetMyTrips(userID uuid.UUID) ([]Trip, error) {
	owned, err := s.repo.FindTripsByOwner(userID)
	if err != nil {
		return nil, err
	}
	joined, err := s.repo.FindTripsByMember(userID)
	if err != nil {
		return nil, err
	}

	seen := make(map[uuid.UUID]bool)
	merged := make([]Trip, 0, len(owned)+len(joined))
	for _, t := range owned {
		if !seen[t.ID] {
			seen[t.ID] = true
			merged = append(merged, t)
		}
	}
	for _, t := range joined {
		if !seen[t.ID] {
			seen[t.ID] = true
			merged = append(merged, t)
		}
	}
	return merged, nil
}

func (s *service) UpdateTrip(tripID, requesterID uuid.UUID, req *UpdateTripRequest) (*Trip, error) {
	trip, err := s.repo.FindTripByID(tripID)
	if err != nil {
		return nil, err
	}
	if !s.isOwnerOrAdmin(tripID, requesterID) {
		return nil, errors.New("only owner or admin can update trip")
	}
	if req.Title != "" {
		trip.Title = req.Title
	}
	if req.Destination != "" {
		trip.Destination = req.Destination
	}
	if req.Status != "" {
		trip.Status = req.Status
	}
	if req.TripType != "" {
		trip.TripType = req.TripType
	}
	if req.VibeTags != nil {
		trip.VibeTags = pq.StringArray(req.VibeTags)
	}
	if req.TravelersPlanned > 0 {
		trip.TravelersPlanned = req.TravelersPlanned
	}
	if req.StartDate != nil {
		trip.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		trip.EndDate = req.EndDate
	}
	if req.Budget > 0 {
		trip.Budget = req.Budget
	}
	if req.CoverImageURL != nil {
		trip.CoverImageURL = req.CoverImageURL
	}
	if req.Notes != nil {
		trip.Notes = req.Notes
	}
	if err := s.repo.UpdateTrip(trip); err != nil {
		return nil, err
	}
	return trip, nil
}

func (s *service) DeleteTrip(tripID, requesterID uuid.UUID) error {
	trip, err := s.repo.FindTripByID(tripID)
	if err != nil {
		return err
	}
	if trip.OwnerUserID != requesterID {
		return errors.New("only the trip owner can delete this trip")
	}
	return s.repo.DeleteTrip(tripID)
}

func (s *service) InviteMember(tripID, requesterID uuid.UUID, req *InviteMemberRequest) (*TripMember, error) {
	if !s.isOwnerOrAdmin(tripID, requesterID) {
		return nil, errors.New("only owner or admin can invite members")
	}
	existing, err := s.repo.FindMember(tripID, req.UserID)
	if err == nil && existing != nil {
		return nil, errors.New("user is already a member or invited")
	}
	role := req.Role
	if role == "" {
		role = RoleMember
	}
	member := &TripMember{
		TripID:     tripID,
		UserID:     req.UserID,
		Role:       role,
		JoinStatus: JoinStatusInvited,
	}
	if err := s.repo.AddMember(member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *service) UpdateMemberStatus(tripID, userID uuid.UUID, req *UpdateMemberStatusRequest) (*TripMember, error) {
	member, err := s.repo.FindMember(tripID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("membership not found")
		}
		return nil, err
	}
	member.JoinStatus = req.JoinStatus
	if req.JoinStatus == JoinStatusJoined {
		now := time.Now()
		member.JoinedAt = &now
	}
	if err := s.repo.UpdateMember(member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *service) RemoveMember(tripID, requesterID, targetUserID uuid.UUID) error {
	if !s.isOwnerOrAdmin(tripID, requesterID) && requesterID != targetUserID {
		return errors.New("not authorized to remove this member")
	}
	return s.repo.RemoveMember(tripID, targetUserID)
}

func (s *service) GetMembers(tripID uuid.UUID) ([]TripMember, error) {
	return s.repo.FindMembersByTrip(tripID)
}

func (s *service) AddItineraryItem(tripID, userID uuid.UUID, req *CreateItineraryItemRequest) (*TripItineraryItem, error) {
	if !s.isMember(tripID, userID) {
		return nil, errors.New("only trip members can add itinerary items")
	}
	peopleCount := req.PeopleCount
	if peopleCount <= 0 {
		peopleCount = 1
	}
	item := &TripItineraryItem{
		TripID:          tripID,
		DayNumber:       req.DayNumber,
		Title:           req.Title,
		ItemType:        req.ItemType,
		PeopleCount:     peopleCount,
		StartTime:       req.StartTime,
		LocationName:    req.LocationName,
		Notes:           req.Notes,
		Lat:             req.Lat,
		Lng:             req.Lng,
		SortOrder:       req.SortOrder,
		CreatedByUserID: &userID,
	}
	if err := s.repo.CreateItineraryItem(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *service) GetItinerary(tripID uuid.UUID) ([]TripItineraryItem, error) {
	return s.repo.FindItineraryByTrip(tripID)
}

func (s *service) UpdateItineraryItem(itemID, requesterID uuid.UUID, req *UpdateItineraryItemRequest) (*TripItineraryItem, error) {
	item, err := s.repo.FindItineraryItemByID(itemID)
	if err != nil {
		return nil, err
	}
	if !s.isMember(item.TripID, requesterID) {
		return nil, errors.New("access denied")
	}
	if req.Title != "" {
		item.Title = req.Title
	}
	if req.ItemType != "" {
		item.ItemType = req.ItemType
	}
	if req.PeopleCount > 0 {
		item.PeopleCount = req.PeopleCount
	}
	if req.DayNumber > 0 {
		item.DayNumber = req.DayNumber
	}
	if req.StartTime != nil {
		item.StartTime = req.StartTime
	}
	if req.LocationName != "" {
		item.LocationName = req.LocationName
	}
	if req.Notes != nil {
		item.Notes = req.Notes
	}
	if req.Lat != nil {
		item.Lat = req.Lat
	}
	if req.Lng != nil {
		item.Lng = req.Lng
	}
	item.SortOrder = req.SortOrder
	if err := s.repo.UpdateItineraryItem(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *service) DeleteItineraryItem(itemID, requesterID uuid.UUID) error {
	item, err := s.repo.FindItineraryItemByID(itemID)
	if err != nil {
		return err
	}
	if !s.isOwnerOrAdmin(item.TripID, requesterID) {
		return errors.New("only owner or admin can delete itinerary items")
	}
	return s.repo.DeleteItineraryItem(itemID)
}

func (s *service) AddExpense(tripID, userID uuid.UUID, req *CreateExpenseRequest) (*TripExpense, error) {
	if !s.isMember(tripID, userID) {
		return nil, errors.New("only trip members can log expenses")
	}
	description := strings.TrimSpace(req.Name)
	if description == "" {
		description = strings.TrimSpace(req.Description)
	}
	if description == "" {
		return nil, errors.New("name is required")
	}

	amount := req.Cost
	if amount <= 0 {
		amount = req.Amount
	}
	if amount <= 0 {
		return nil, errors.New("cost is required")
	}

	currency := req.CurrencyCode
	if currency == "" {
		currency = "USD"
	}
	expense := &TripExpense{
		TripID:          tripID,
		CreatedByUserID: &userID,
		Description:     description,
		LocationName:    strings.TrimSpace(req.LocationName),
		Category:        req.Category,
		Amount:          amount,
		ExpenseDate:     req.ExpenseDate,
		CurrencyCode:    currency,
	}
	if err := s.repo.CreateExpense(expense); err != nil {
		return nil, err
	}

	total, _ := s.repo.SumExpensesByTrip(tripID)
	trip, err := s.repo.FindTripByID(tripID)
	if err == nil {
		trip.Spent = total
		_ = s.repo.UpdateTrip(trip)
	}
	return expense, nil
}

func (s *service) GetExpenses(tripID uuid.UUID) ([]TripExpense, error) {
	return s.repo.FindExpensesByTrip(tripID)
}

func (s *service) UpdateExpense(expenseID, requesterID uuid.UUID, req *UpdateExpenseRequest) (*TripExpense, error) {
	expense, err := s.repo.FindExpenseByID(expenseID)
	if err != nil {
		return nil, err
	}
	if expense.CreatedByUserID == nil || *expense.CreatedByUserID != requesterID {
		if !s.isOwnerOrAdmin(expense.TripID, requesterID) {
			return nil, errors.New("not authorized to edit this expense")
		}
	}
	if strings.TrimSpace(req.Name) != "" {
		expense.Description = strings.TrimSpace(req.Name)
	} else if strings.TrimSpace(req.Description) != "" {
		expense.Description = strings.TrimSpace(req.Description)
	}
	if strings.TrimSpace(req.LocationName) != "" {
		expense.LocationName = strings.TrimSpace(req.LocationName)
	}
	if req.Category != "" {
		expense.Category = req.Category
	}
	if req.Cost > 0 {
		expense.Amount = req.Cost
	} else if req.Amount > 0 {
		expense.Amount = req.Amount
	}
	if !req.ExpenseDate.IsZero() {
		expense.ExpenseDate = req.ExpenseDate
	}
	if req.CurrencyCode != "" {
		expense.CurrencyCode = req.CurrencyCode
	}
	if err := s.repo.UpdateExpense(expense); err != nil {
		return nil, err
	}

	total, _ := s.repo.SumExpensesByTrip(expense.TripID)
	trip, err := s.repo.FindTripByID(expense.TripID)
	if err == nil {
		trip.Spent = total
		_ = s.repo.UpdateTrip(trip)
	}
	return expense, nil
}

func (s *service) DeleteExpense(expenseID, requesterID uuid.UUID) error {
	expense, err := s.repo.FindExpenseByID(expenseID)
	if err != nil {
		return err
	}
	if expense.CreatedByUserID == nil || *expense.CreatedByUserID != requesterID {
		if !s.isOwnerOrAdmin(expense.TripID, requesterID) {
			return errors.New("not authorized to delete this expense")
		}
	}
	if err := s.repo.DeleteExpense(expenseID); err != nil {
		return err
	}

	total, _ := s.repo.SumExpensesByTrip(expense.TripID)
	trip, err := s.repo.FindTripByID(expense.TripID)
	if err == nil {
		trip.Spent = total
		_ = s.repo.UpdateTrip(trip)
	}
	return nil
}

type aiGeneratedActivity struct {
	DayNumber    int      `json:"day_number"`
	Title        string   `json:"title"`
	ItemType     ItemType `json:"item_type"`
	PeopleCount  int      `json:"people_count"`
	StartTime    string   `json:"start_time"`
	LocationName string   `json:"location_name"`
	Notes        string   `json:"notes"`
}

func (s *service) GenerateAndSaveAIItinerary(tripID, userID uuid.UUID, req *GenerateAIItineraryRequest) ([]TripItineraryItem, error) {
	if !s.isMember(tripID, userID) {
		return nil, errors.New("only trip members can generate itinerary")
	}
	if req.EndDate.Before(req.StartDate) {
		return nil, errors.New("end_date must be after start_date")
	}

	activities, err := s.generateActivities(req)
	if err != nil {
		return nil, err
	}

	created := make([]TripItineraryItem, 0, len(activities))
	for i, a := range activities {
		if strings.TrimSpace(a.Title) == "" {
			continue
		}

		dayNumber := a.DayNumber
		if dayNumber <= 0 {
			dayNumber = 1
		}

		people := a.PeopleCount
		if people <= 0 {
			people = req.NumberOfPeople
			if people <= 0 {
				people = 1
			}
		}

		itemType := a.ItemType
		if itemType == "" {
			itemType = ItemTypeActivity
		}

		var startTime *time.Time
		if a.StartTime != "" {
			if t, parseErr := time.Parse(time.RFC3339, a.StartTime); parseErr == nil {
				startTime = &t
			}
		}

		notes := strings.TrimSpace(a.Notes)
		var notesPtr *string
		if notes != "" {
			notesCopy := notes
			notesPtr = &notesCopy
		}

		item := &TripItineraryItem{
			TripID:          tripID,
			DayNumber:       dayNumber,
			Title:           strings.TrimSpace(a.Title),
			ItemType:        itemType,
			PeopleCount:     people,
			StartTime:       startTime,
			LocationName:    strings.TrimSpace(a.LocationName),
			Notes:           notesPtr,
			SortOrder:       i + 1,
			CreatedByUserID: &userID,
		}

		if err := s.repo.CreateItineraryItem(item); err != nil {
			return nil, err
		}
		created = append(created, *item)
	}

	return created, nil
}

func (s *service) PlanAndCreateTripWithAI(userID uuid.UUID, req *PlanAndCreateTripRequest) (*PlanAndCreateTripResponse, error) {
	if req.EndDate.Before(req.StartDate) {
		return nil, errors.New("end_date must be after start_date")
	}

	tripTitle := strings.TrimSpace(req.Title)
	if tripTitle == "" {
		tripTitle = strings.TrimSpace(req.Location) + " Trip"
	}

	start := req.StartDate
	end := req.EndDate
	createReq := &CreateTripRequest{
		Title:            tripTitle,
		Destination:      req.Location,
		TravelersPlanned: req.NumberOfPeople,
		StartDate:        &start,
		EndDate:          &end,
		Budget:           req.Budget,
	}

	if strings.TrimSpace(req.Vibe) != "" {
		createReq.VibeTags = []string{strings.TrimSpace(req.Vibe)}
	}

	trip, err := s.CreateTrip(userID, createReq)
	if err != nil {
		return nil, err
	}

	itReq := &GenerateAIItineraryRequest{
		Location:       req.Location,
		Vibe:           req.Vibe,
		NumberOfPeople: req.NumberOfPeople,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		Budget:         req.Budget,
		Prompt:         req.Prompt,
	}

	items, err := s.GenerateAndSaveAIItinerary(trip.ID, userID, itReq)
	if err != nil {
		return nil, err
	}

	return &PlanAndCreateTripResponse{
		Trip:      trip,
		Itinerary: items,
	}, nil
}

func (s *service) generateActivities(req *GenerateAIItineraryRequest) ([]aiGeneratedActivity, error) {
	if s.grokKey == "" {
		return fallbackActivities(req), nil
	}

	prompt := fmt.Sprintf("Generate a detailed travel itinerary as JSON array only. Location: %s. Vibe: %s. People: %d. Start: %s. End: %s. Budget: %.2f. %s\nEach item must include keys: day_number,title,item_type,people_count,start_time,location_name,notes. start_time must be RFC3339.",
		req.Location,
		req.Vibe,
		req.NumberOfPeople,
		req.StartDate.Format(time.RFC3339),
		req.EndDate.Format(time.RFC3339),
		req.Budget,
		req.Prompt,
	)

	bodyMap := map[string]interface{}{
		"model": "grok-beta",
		"messages": []map[string]string{{
			"role":    "user",
			"content": prompt,
		}},
	}

	body, _ := json.Marshal(bodyMap)
	httpReq, err := http.NewRequest("POST", "https://api.x.ai/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.grokKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fallbackActivities(req), nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var grokResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &grokResp); err != nil || len(grokResp.Choices) == 0 {
		return fallbackActivities(req), nil
	}

	content := strings.TrimSpace(grokResp.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var activities []aiGeneratedActivity
	if err := json.Unmarshal([]byte(content), &activities); err != nil || len(activities) == 0 {
		return fallbackActivities(req), nil
	}

	return activities, nil
}

func fallbackActivities(req *GenerateAIItineraryRequest) []aiGeneratedActivity {
	people := req.NumberOfPeople
	if people <= 0 {
		people = 1
	}

	days := int(req.EndDate.Sub(req.StartDate).Hours()/24) + 1
	if days < 1 {
		days = 1
	}
	if days > 10 {
		days = 10
	}

	activities := make([]aiGeneratedActivity, 0, days*3)
	for d := 1; d <= days; d++ {
		baseDate := req.StartDate.AddDate(0, 0, d-1)
		morning := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), 9, 0, 0, 0, time.UTC)
		afternoon := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), 13, 0, 0, 0, time.UTC)
		evening := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), 18, 0, 0, 0, time.UTC)

		activities = append(activities,
			aiGeneratedActivity{DayNumber: d, Title: "Morning exploration", ItemType: ItemTypeActivity, PeopleCount: people, StartTime: morning.Format(time.RFC3339), LocationName: req.Location, Notes: "Walk key neighborhoods and landmarks."},
			aiGeneratedActivity{DayNumber: d, Title: "Lunch stop", ItemType: ItemTypeFood, PeopleCount: people, StartTime: afternoon.Format(time.RFC3339), LocationName: req.Location, Notes: "Try highly rated local cuisine."},
			aiGeneratedActivity{DayNumber: d, Title: "Evening highlight", ItemType: ItemTypeActivity, PeopleCount: people, StartTime: evening.Format(time.RFC3339), LocationName: req.Location, Notes: "Sunset activity and dinner area."},
		)
	}

	return activities
}

func (s *service) GetChatHistory(tripID uuid.UUID, limit int) ([]ChatMessage, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	return s.repo.FindChatMessages(tripID, limit)
}

func (s *service) SendChatMessage(tripID, userID uuid.UUID, message string) (*ChatMessage, error) {
	if message == "" {
		return nil, errors.New("message cannot be empty")
	}
	// Any joined member can send a message — owner check not required for chat.
	if !s.isMember(tripID, userID) {
		return nil, errors.New("only trip members can send messages")
	}
	msg := &ChatMessage{
		TripID:  tripID,
		UserID:  userID,
		Message: message,
	}
	if err := s.repo.CreateChatMessage(msg); err != nil {
		return nil, err
	}
	return msg, nil
}
