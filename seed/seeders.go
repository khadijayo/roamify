package seed

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"roamify/internal/models"
)

func Run(db *gorm.DB) {
	log.Println("🌱 Starting seed...")

	seedWilayas(db)
	seedDestinations(db)
	seedHotels(db)
	seedUsers(db)

	log.Println("✅ Seed complete.")
}


func seedWilayas(db *gorm.DB) {
	wilayas := []models.Wilaya{
		{Code: 1,  Name: "Adrar",         ArabicName: "أدرار",       Latitude: 27.87, Longitude: -0.29},
		{Code: 6,  Name: "Béjaïa",        ArabicName: "بجاية",       Latitude: 36.75, Longitude: 5.08},
		{Code: 16, Name: "Algiers",        ArabicName: "الجزائر",     Latitude: 36.74, Longitude: 3.06,  CoverImage: "https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=800"},
		{Code: 13, Name: "Tlemcen",        ArabicName: "تلمسان",      Latitude: 34.87, Longitude: -1.31, CoverImage: "https://images.unsplash.com/photo-1539650116574-75c0c6d73f6e?w=800"},
		{Code: 21, Name: "Skikda",         ArabicName: "سكيكدة",      Latitude: 36.87, Longitude: 6.91},
		{Code: 23, Name: "Annaba",         ArabicName: "عنابة",       Latitude: 36.91, Longitude: 7.76},
		{Code: 25, Name: "Constantine",    ArabicName: "قسنطينة",     Latitude: 36.36, Longitude: 6.61,  CoverImage: "https://images.unsplash.com/photo-1590736969596-f4b3b0c1b0c8?w=800"},
		{Code: 30, Name: "Ouargla",        ArabicName: "ورقلة",       Latitude: 31.95, Longitude: 5.32},
		{Code: 44, Name: "Aïn Temouchent", ArabicName: "عين تموشنت",  Latitude: 35.30, Longitude: -1.14},
		{Code: 11, Name: "Tamanrasset",    ArabicName: "تمنراست",     Latitude: 22.78, Longitude: 5.52,  CoverImage: "https://images.unsplash.com/photo-1509316785289-025f5b846b35?w=800"},
		{Code: 8,  Name: "Béchar",         ArabicName: "بشار",        Latitude: 31.62, Longitude: -2.22},
		{Code: 31, Name: "Oran",           ArabicName: "وهران",       Latitude: 35.69, Longitude: -0.63, CoverImage: "https://images.unsplash.com/photo-1568454537842-d933259bb258?w=800"},
		{Code: 56, Name: "Djanet",         ArabicName: "جانت",        Latitude: 24.55, Longitude: 9.48},
		{Code: 49, Name: "Timimoun",       ArabicName: "تيميمون",     Latitude: 29.26, Longitude: 0.23},
		{Code: 5,  Name: "Batna",          ArabicName: "باتنة",       Latitude: 35.55, Longitude: 6.17},
		{Code: 19, Name: "Sétif",          ArabicName: "سطيف",        Latitude: 36.19, Longitude: 5.41},
		{Code: 15, Name: "Tizi Ouzou",     ArabicName: "تيزي وزو",    Latitude: 36.71, Longitude: 4.05},
	}

	for _, w := range wilayas {
		db.FirstOrCreate(&w, models.Wilaya{Code: w.Code})
	}
	log.Println("  ✓ Wilayas seeded")
}


func boolPtr(b bool) *bool { return &b }

func getWilayaID(db *gorm.DB, name string) uint {
	var w models.Wilaya
	db.Where("name = ?", name).First(&w)
	return w.ID
}


func seedDestinations(db *gorm.DB) {
	type destSeed struct {
		place  models.Place
		detail models.DestinationDetail
		images []string
		acts   []models.ActivityType
		seasons []models.Season
	}

	algiersID   := getWilayaID(db, "Algiers")
	tlemcenID   := getWilayaID(db, "Tlemcen")
	annabaID    := getWilayaID(db, "Annaba")
	tamanID     := getWilayaID(db, "Tamanrasset")
	bejaiID     := getWilayaID(db, "Béjaïa")

	destinations := []destSeed{
		{
			place: models.Place{
				Name: "The Casbah", Slug: "the-casbah",
				Category: models.CategoryDestination,
				Description: "UNESCO World Heritage Site and the historic heart of Algiers. A labyrinth of narrow streets, Ottoman palaces, and traditional houses cascading down to the sea.",
				WilayaID: algiersID, Address: "Casbah, Algiers",
				Latitude: 36.7878, Longitude: 3.0598,
				MapsURL: "https://maps.google.com/?q=Casbah+Algiers",
				IsOpen: boolPtr(true), IsFeatured: true,
				AverageRating: 4.5, ReviewCount: 312,
			},
			detail: models.DestinationDetail{
				Type: models.TypeHistorical, EntryFee: 0,
				VisitDuration: "2–3 hours",
				TipsText: "Visit in the morning to avoid crowds. Hire a local guide for the best experience.",
			},
			images: []string{
				"https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=600",
				"https://images.unsplash.com/photo-1578662996442-48f60103fc96?w=600",
			},
			acts:    []models.ActivityType{models.ActivityCultural, models.ActivityShopping},
			seasons: []models.Season{models.SeasonSpring, models.SeasonAutumn},
		},
		{
			place: models.Place{
				Name: "Cap Carbon", Slug: "cap-carbon",
				Category: models.CategoryDestination,
				Description: "A stunning cape on the Mediterranean coast of Béjaïa, known for its lighthouse, dramatic cliffs, and crystal-clear waters. A paradise for nature lovers.",
				WilayaID: bejaiID, Address: "Cap Carbon, Béjaïa",
				Latitude: 36.7786, Longitude: 5.1147,
				MapsURL: "https://maps.google.com/?q=Cap+Carbon+Bejaia",
				IsOpen: boolPtr(true), IsFeatured: true,
				AverageRating: 4.7, ReviewCount: 189,
			},
			detail: models.DestinationDetail{
				Type: models.TypeBeach, EntryFee: 0,
				VisitDuration: "Half day",
				TipsText: "Bring water and sunscreen. The hike to the lighthouse is worth it for panoramic views.",
			},
			images: []string{
				"https://images.unsplash.com/photo-1507525428034-b723cf961d3e?w=600",
			},
			acts:    []models.ActivityType{models.ActivityHiking, models.ActivitySwimming},
			seasons: []models.Season{models.SeasonSummer, models.SeasonSpring},
		},
		{
			place: models.Place{
				Name: "Tlemcen Great Mosque", Slug: "tlemcen-great-mosque",
				Category: models.CategoryDestination,
				Description: "One of Algeria's most magnificent examples of Almoravid architecture, dating back to 1082. Features an intricate minaret and stunning tilework.",
				WilayaID: tlemcenID, Address: "Old Town, Tlemcen",
				Latitude: 34.8784, Longitude: -1.3152,
				MapsURL: "https://maps.google.com/?q=Great+Mosque+Tlemcen",
				IsOpen: boolPtr(true), IsFeatured: false,
				AverageRating: 4.6, ReviewCount: 145,
			},
			detail: models.DestinationDetail{
				Type: models.TypeCulturalSite, EntryFee: 0,
				VisitDuration: "1 hour",
				TipsText: "Dress modestly. Visit on Friday to experience the atmosphere, but respect prayer times.",
			},
			images: []string{
				"https://images.unsplash.com/photo-1539650116574-75c0c6d73f6e?w=600",
			},
			acts:    []models.ActivityType{models.ActivityCultural},
			seasons: []models.Season{models.SeasonSpring, models.SeasonAutumn, models.SeasonWinter},
		},
		{
			place: models.Place{
				Name: "Sidi Fredj Beach", Slug: "sidi-fredj-beach",
				Category: models.CategoryDestination,
				Description: "A popular coastal resort west of Algiers with a beautiful marina, sandy beaches, restaurants, and vibrant nightlife.",
				WilayaID: algiersID, Address: "Sidi Fredj, Algiers",
				Latitude: 36.7386, Longitude: 2.8755,
				MapsURL: "https://maps.google.com/?q=Sidi+Fredj+Beach",
				IsOpen: boolPtr(true), IsFeatured: false,
				AverageRating: 4.1, ReviewCount: 278,
			},
			detail: models.DestinationDetail{
				Type: models.TypeBeach, EntryFee: 0,
				VisitDuration: "Full day",
				TipsText: "Weekdays are much less crowded. The marina area has great seafood restaurants.",
			},
			images: []string{
				"https://images.unsplash.com/photo-1507525428034-b723cf961d3e?w=600",
			},
			acts:    []models.ActivityType{models.ActivitySwimming, models.ActivityRelaxation, models.ActivityFoodExperience},
			seasons: []models.Season{models.SeasonSummer},
		},
		{
			place: models.Place{
				Name: "Seraïdi Forest", Slug: "seraidi-forest",
				Category: models.CategoryDestination,
				Description: "A lush mountain village surrounded by dense cork oak forests near Annaba. Stunning panoramic views over the Mediterranean.",
				WilayaID: annabaID, Address: "Seraïdi, Annaba",
				Latitude: 36.9592, Longitude: 7.7139,
				MapsURL: "https://maps.google.com/?q=Seraidi+Annaba",
				IsOpen: boolPtr(true), IsFeatured: false,
				AverageRating: 4.4, ReviewCount: 97,
			},
			detail: models.DestinationDetail{
				Type: models.TypeMountain, EntryFee: 0,
				VisitDuration: "Full day",
				TipsText: "Great for weekend escape from Annaba city. Bring a jacket even in summer.",
			},
			images: []string{
				"https://images.unsplash.com/photo-1448375240586-882707db888b?w=600",
			},
			acts:    []models.ActivityType{models.ActivityHiking, models.ActivityCamping, models.ActivityRelaxation},
			seasons: []models.Season{models.SeasonSpring, models.SeasonSummer, models.SeasonAutumn},
		},
		{
			place: models.Place{
				Name: "Hoggar Mountains", Slug: "hoggar-mountains",
				Category: models.CategoryDestination,
				Description: "An otherworldly volcanic massif in the heart of the Sahara. Ancient rock art, dramatic landscapes, and the hermitage of Charles de Foucauld atop Assekrem.",
				WilayaID: tamanID, Address: "Hoggar, Tamanrasset",
				Latitude: 23.2833, Longitude: 5.5667,
				MapsURL: "https://maps.google.com/?q=Hoggar+Mountains+Algeria",
				IsOpen: boolPtr(true), IsFeatured: true,
				AverageRating: 4.9, ReviewCount: 211,
			},
			detail: models.DestinationDetail{
				Type: models.TypeDesert, EntryFee: 0,
				VisitDuration: "2–5 days",
				Accessibility: "4x4 vehicle required. Guided tours recommended.",
				TipsText: "Visit October–March to avoid extreme heat. Sunrise from Assekrem is unforgettable.",
			},
			images: []string{
				"https://images.unsplash.com/photo-1509316785289-025f5b846b35?w=600",
				"https://images.unsplash.com/photo-1509316785289-025f5b846b35?w=600",
			},
			acts:    []models.ActivityType{models.ActivityAdventure, models.ActivityCamping, models.ActivityHiking, models.ActivityCultural},
			seasons: []models.Season{models.SeasonAutumn, models.SeasonWinter, models.SeasonSpring},
		},
	}

	for _, d := range destinations {
		d.place.ID = uuid.New()

		var existing models.Place
		if err := db.Where("slug = ?", d.place.Slug).First(&existing).Error; err == nil {
			continue // already seeded
		}

		if err := db.Create(&d.place).Error; err != nil {
			log.Printf("  ✗ Failed to create destination %s: %v", d.place.Name, err)
			continue
		}

		d.detail.PlaceID = d.place.ID
		db.Create(&d.detail)

		for i, url := range d.images {
			db.Create(&models.PlaceImage{
				PlaceID:   d.place.ID,
				URL:       url,
				IsCover:   i == 0,
				SortOrder: i,
				AltText:   d.place.Name,
			})
		}

		for _, act := range d.acts {
			db.Create(&models.PlaceActivity{PlaceID: d.place.ID, Activity: act})
		}

		for _, s := range d.seasons {
			db.Create(&models.PlaceSeason{PlaceID: d.place.ID, Season: s})
		}
	}

	log.Println("  ✓ Destinations seeded")
}


func seedHotels(db *gorm.DB) {
	type hotelSeed struct {
		place  models.Place
		detail models.HotelDetail
		image  string
	}

	algiersID    := getWilayaID(db, "Algiers")
	constantineID := getWilayaID(db, "Constantine")
	annabaID     := getWilayaID(db, "Annaba")
	tlemcenID    := getWilayaID(db, "Tlemcen")
	tamanID      := getWilayaID(db, "Tamanrasset")

	hotels := []hotelSeed{
		{
			place: models.Place{
				Name: "El-Auressi Hotel", Slug: "el-auressi-hotel",
				Category: models.CategoryHotel,
				Description: "A landmark 5-star hotel in the heart of Algiers with panoramic views of the Bay of Algiers.",
				WilayaID: algiersID, Address: "1 Rue Asselah Hocine, Algiers",
				Latitude: 36.7400, Longitude: 3.0600,
				MapsURL: "https://maps.google.com/?q=El+Auressi+Hotel+Algiers",
				IsOpen: boolPtr(true), IsFeatured: true,
				AverageRating: 4.3, ReviewCount: 456,
			},
			detail: models.HotelDetail{
				StarRating: 5, PricePerNight: 18000,
				PhoneNumber: "+213 21 73 73 73", Email: "info@elauressi.dz",
				CheckIn: "14:00", CheckOut: "12:00",
				HasPool: true, HasWifi: true, HasParking: true, HasRestaurant: true,
			},
			image: "https://images.unsplash.com/photo-1566073771259-6a8506099945?w=600",
		},
		{
			place: models.Place{
				Name: "Royal Hotel", Slug: "royal-hotel-constantine",
				Category: models.CategoryHotel,
				Description: "Elegant hotel offering stunning views of the famous Rhumel Gorge and modern amenities in the heart of Constantine.",
				WilayaID: constantineID, Address: "Boulevard Zighoud Youcef, Constantine",
				Latitude: 36.3650, Longitude: 6.6147,
				MapsURL: "https://maps.google.com/?q=Royal+Hotel+Constantine",
				IsOpen: boolPtr(true), IsFeatured: false,
				AverageRating: 4.1, ReviewCount: 203,
			},
			detail: models.HotelDetail{
				StarRating: 4, PricePerNight: 9500,
				PhoneNumber: "+213 31 93 15 00",
				CheckIn: "15:00", CheckOut: "11:00",
				HasPool: false, HasWifi: true, HasParking: true, HasRestaurant: true,
			},
			image: "https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?w=600",
		},
		{
			place: models.Place{
				Name: "Marriott Hotel", Slug: "marriott-constantine",
				Category: models.CategoryHotel,
				Description: "International standard 5-star hotel in Constantine with world-class facilities and exceptional service.",
				WilayaID: constantineID, Address: "Constantine Business Park",
				Latitude: 36.3700, Longitude: 6.6200,
				MapsURL: "https://maps.google.com/?q=Marriott+Constantine+Algeria",
				IsOpen: boolPtr(true), IsFeatured: true,
				AverageRating: 4.6, ReviewCount: 378,
			},
			detail: models.HotelDetail{
				StarRating: 5, PricePerNight: 22000,
				PhoneNumber: "+213 31 94 50 00", Website: "https://marriott.com",
				CheckIn: "15:00", CheckOut: "12:00",
				HasPool: true, HasWifi: true, HasParking: true, HasRestaurant: true,
			},
			image: "https://images.unsplash.com/photo-1582719508461-905c673771fd?w=600",
		},
		{
			place: models.Place{
				Name: "Camp Dunes D'Or", Slug: "camp-dunes-dor",
				Category: models.CategoryHotel,
				Description: "Unique desert glamping experience in the heart of the Sahara. Luxury tents, camel rides, and stargazing under the Algerian sky.",
				WilayaID: tamanID, Address: "Tamanrasset Desert Camp",
				Latitude: 22.7900, Longitude: 5.5300,
				MapsURL: "https://maps.google.com/?q=Tamanrasset+Desert+Camp",
				IsOpen: boolPtr(true), IsFeatured: true,
				AverageRating: 4.8, ReviewCount: 124,
			},
			detail: models.HotelDetail{
				StarRating: 3, PricePerNight: 12000,
				PhoneNumber: "+213 29 34 12 34",
				CheckIn: "16:00", CheckOut: "10:00",
				HasPool: false, HasWifi: false, HasParking: true, HasRestaurant: true,
			},
			image: "https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=600",
		},
		{
			place: models.Place{
				Name: "Hôtel Zianides", Slug: "hotel-zianides-tlemcen",
				Category: models.CategoryHotel,
				Description: "Boutique hotel in Tlemcen blending traditional Andalusian architecture with modern comfort, near the Great Mosque.",
				WilayaID: tlemcenID, Address: "Centre Ville, Tlemcen",
				Latitude: 34.8780, Longitude: -1.3140,
				MapsURL: "https://maps.google.com/?q=Hotel+Zianides+Tlemcen",
				IsOpen: boolPtr(true), IsFeatured: false,
				AverageRating: 4.2, ReviewCount: 89,
			},
			detail: models.HotelDetail{
				StarRating: 4, PricePerNight: 8500,
				PhoneNumber: "+213 43 20 15 00",
				CheckIn: "14:00", CheckOut: "12:00",
				HasPool: true, HasWifi: true, HasParking: false, HasRestaurant: true,
			},
			image: "https://images.unsplash.com/photo-1564501049412-61c2a3083791?w=600",
		},
		{
			place: models.Place{
				Name: "Sheraton Annaba", Slug: "sheraton-annaba",
				Category: models.CategoryHotel,
				Description: "Luxury beachfront hotel in Annaba with direct Mediterranean access and a full spa.",
				WilayaID: annabaID, Address: "Plage de Richelieu, Annaba",
				Latitude: 36.9150, Longitude: 7.7550,
				MapsURL: "https://maps.google.com/?q=Sheraton+Annaba",
				IsOpen: boolPtr(true), IsFeatured: false,
				AverageRating: 4.5, ReviewCount: 267,
			},
			detail: models.HotelDetail{
				StarRating: 5, PricePerNight: 20000,
				PhoneNumber: "+213 38 86 00 00", Website: "https://sheraton.com",
				CheckIn: "15:00", CheckOut: "12:00",
				HasPool: true, HasWifi: true, HasParking: true, HasRestaurant: true,
			},
			image: "https://images.unsplash.com/photo-1520250497591-112f2f40a3f4?w=600",
		},
	}

	for _, h := range hotels {
		h.place.ID = uuid.New()

		var existing models.Place
		if err := db.Where("slug = ?", h.place.Slug).First(&existing).Error; err == nil {
			continue
		}

		if err := db.Create(&h.place).Error; err != nil {
			log.Printf("  ✗ Failed to create hotel %s: %v", h.place.Name, err)
			continue
		}

		h.detail.PlaceID = h.place.ID
		db.Create(&h.detail)

		db.Create(&models.PlaceImage{
			PlaceID: h.place.ID, URL: h.image,
			IsCover: true, AltText: h.place.Name,
		})
	}

	log.Println("  ✓ Hotels seeded")
}


func seedUsers(db *gorm.DB) {
	now := time.Now()
	users := []models.User{
		{
			ID: uuid.New(), FullName: "Admin Roamify",
			Email: "admin@roamify.dz",
			// NOTE: In production, hash this with bcrypt before storing
			PasswordHash: "$2a$10$placeholder_hash_replace_me",
			IsActive: true, LastLoginAt: &now,
		},
		{
			ID: uuid.New(), FullName: "Youcef Benali",
			Email: "youcef@example.dz",
			PasswordHash: "$2a$10$placeholder_hash_replace_me",
			IsActive: true,
		},
		{
			ID: uuid.New(), FullName: "Amina Khelil",
			Email: "amina@example.dz",
			PasswordHash: "$2a$10$placeholder_hash_replace_me",
			IsActive: true,
		},
	}

	for _, u := range users {
		db.FirstOrCreate(&u, models.User{Email: u.Email})
	}

	log.Println("  ✓ Users seeded")
}