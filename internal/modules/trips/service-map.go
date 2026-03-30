package trips

import "github.com/google/uuid"

// MapPin is the shape returned by GET /trips/:tripId/map.
// Only items with lat+lng are included — items without coordinates are skipped.
type MapPin struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	LocationName string    `json:"location_name"`
	ItemType     ItemType  `json:"item_type"`
	DayNumber    int       `json:"day_number"`
	Lat          float64   `json:"lat"`
	Lng          float64   `json:"lng"`
	
}

// Add GetTripMapPins to your existing Service interface in service.go:
//   GetTripMapPins(tripID uuid.UUID) ([]MapPin, error)

// --- Implementation to add to service struct ---

func (s *service) GetTripMapPins(tripID uuid.UUID) ([]MapPin, error) {
	items, err := s.repo.FindItineraryByTrip(tripID)
	if err != nil {
		return nil, err
	}
	pins := make([]MapPin, 0, len(items))
	for _, item := range items {
		// Skip items with no coordinates
		if item.Lat == nil || item.Lng == nil {
			continue
		}
		pins = append(pins, MapPin{
			ID:           item.ID,
			Title:        item.Title,
			LocationName: item.LocationName,
			ItemType:     item.ItemType,
			DayNumber:    item.DayNumber,
			Lat:          *item.Lat,
			Lng:          *item.Lng,
		})
	}
	return pins, nil
}
