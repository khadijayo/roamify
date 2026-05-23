package trips

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	trips := r.Group("/trips", auth)
	{
		trips.POST("/plan-and-create", h.PlanAndCreateTripWithAI)
		trips.POST("/", h.CreateTrip)
		trips.GET("/", h.GetMyTrips)
		trips.GET("/all", h.GetAllTrips)
		trips.GET("/:tripId", h.GetTrip)
		trips.PATCH("/:tripId", h.UpdateTrip)
		trips.DELETE("/:tripId", h.DeleteTrip)
		trips.POST("/:tripId/join", h.JoinTrip)

		trips.POST("/:tripId/members", h.InviteMember)
		trips.GET("/:tripId/members", h.GetMembers)
		trips.PATCH("/:tripId/members/status", h.UpdateMemberStatus)
		trips.DELETE("/:tripId/members/:userId", h.RemoveMember)

		trips.POST("/:tripId/itinerary", h.AddItineraryItem)
		trips.POST("/:tripId/itinerary/generate-ai", h.GenerateAIItinerary)
		trips.GET("/:tripId/itinerary", h.GetItinerary)
		trips.PATCH("/:tripId/itinerary/:itemId", h.UpdateItineraryItem)
		trips.DELETE("/:tripId/itinerary/:itemId", h.DeleteItineraryItem)

		trips.POST("/:tripId/expenses", h.AddExpense)
		trips.GET("/:tripId/expenses", h.GetExpenses)
		trips.PATCH("/:tripId/expenses/:expenseId", h.UpdateExpense)
		trips.DELETE("/:tripId/expenses/:expenseId", h.DeleteExpense)

		// Squad chat — canonical /messages endpoints
		trips.GET("/:tripId/messages", h.GetChatHistory)
		trips.POST("/:tripId/messages", h.SendChatMessage)

		// /chat kept for backward compatibility
		trips.GET("/:tripId/chat", h.GetChatHistory)
		trips.POST("/:tripId/chat", h.SendChatMessage)

		trips.GET("/:tripId/map", h.GetTripMapPins)
	}
}