package trips

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	trips := rg.Group("/trips", middleware.Auth())
	{
		trips.POST("", h.CreateTrip)
		trips.GET("", h.GetMyTrips)
		trips.GET("/:tripId", h.GetTrip)
		trips.PATCH("/:tripId", h.UpdateTrip)
		trips.DELETE("/:tripId", h.DeleteTrip)

		// Members
		trips.POST("/:tripId/members", h.InviteMember)
		trips.GET("/:tripId/members", h.GetMembers)
		trips.PATCH("/:tripId/members/status", h.UpdateMemberStatus)
		trips.DELETE("/:tripId/members/:userId", h.RemoveMember)

		// Itinerary
		trips.POST("/:tripId/itinerary", h.AddItineraryItem)
		trips.GET("/:tripId/itinerary", h.GetItinerary)
		trips.PATCH("/:tripId/itinerary/:itemId", h.UpdateItineraryItem)
		trips.DELETE("/:tripId/itinerary/:itemId", h.DeleteItineraryItem)

		// Expenses
		trips.POST("/:tripId/expenses", h.AddExpense)
		trips.GET("/:tripId/expenses", h.GetExpenses)
		trips.PATCH("/:tripId/expenses/:expenseId", h.UpdateExpense)
		trips.DELETE("/:tripId/expenses/:expenseId", h.DeleteExpense)
	}
}