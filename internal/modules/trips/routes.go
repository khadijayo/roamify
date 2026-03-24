package trips

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	trips := r.Group("/trips", auth)
	{
		trips.POST("/", h.CreateTrip)
		trips.GET("/", h.GetMyTrips)
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