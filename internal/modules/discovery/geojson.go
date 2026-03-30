package discovery

// GeoJSON types for the atlas map layer.
// The frontend passes this directly to Mapbox / Google Maps as a layer source.

type GeoJSONFeatureCollection struct {
	Type     string          `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"` // [lng, lat] — GeoJSON order
}

// GetAtlasGeoJSON converts all seeded discovery locations into a GeoJSON
// FeatureCollection that Mapbox / Google Maps can render directly as a layer.
func GetAtlasGeoJSON() *GeoJSONFeatureCollection {
	features := make([]GeoJSONFeature, 0, len(seededLocations))
	for _, loc := range seededLocations {
		features = append(features, GeoJSONFeature{
			Type: "Feature",
			Geometry: GeoJSONGeometry{
				Type:        "Point",
				Coordinates: []float64{loc.Lng, loc.Lat}, // GeoJSON is [lng, lat]
			},
			Properties: map[string]interface{}{
				"id":              loc.ID.String(),
				"name":            loc.Name,
				"country":         loc.Country,
				"vibe":            loc.Vibe,
				"category":        loc.Category,
				"price_from":      loc.PriceFrom,
				"hidden_gem":      loc.HiddenGem,
				"instagrammable":  loc.Instagrammable,
				"description":     loc.Description,
				"image_url":       loc.ImageURL,
				"booking_url":     loc.BookingURL,
			},
		})
	}
	return &GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}
