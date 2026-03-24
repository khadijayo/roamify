package trips

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khadijayo/roamify/pkg/middleware"
	"github.com/khadijayo/roamify/pkg/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ---- Trips ----

// POST /trips
func (h *Handler) CreateTrip(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req CreateTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	trip, err := h.svc.CreateTrip(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, "trip created", trip)
}

// GET /trips
func (h *Handler) GetMyTrips(c *gin.Context) {
	userID := middleware.GetUserID(c)
	trips, err := h.svc.GetMyTrips(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "trips fetched", trips)
}

// GET /trips/:tripId
func (h *Handler) GetTrip(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}
	trip, err := h.svc.GetTrip(tripID, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, "trip fetched", trip)
}

// PATCH /trips/:tripId
func (h *Handler) UpdateTrip(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}
	var req UpdateTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	trip, err := h.svc.UpdateTrip(tripID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "trip updated", trip)
}

// DELETE /trips/:tripId
func (h *Handler) DeleteTrip(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}
	if err := h.svc.DeleteTrip(tripID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "trip deleted", nil)
}

// ---- Members ----

// POST /trips/:tripId/members
func (h *Handler) InviteMember(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}
	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	member, err := h.svc.InviteMember(tripID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "member invited", member)
}

// GET /trips/:tripId/members
func (h *Handler) GetMembers(c *gin.Context) {
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}
	members, err := h.svc.GetMembers(tripID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "members fetched", members)
}

// PATCH /trips/:tripId/members/status
func (h *Handler) UpdateMemberStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}
	var req UpdateMemberStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	member, err := h.svc.UpdateMemberStatus(tripID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "membership status updated", member)
}

// DELETE /trips/:tripId/members/:userId
func (h *Handler) RemoveMember(c *gin.Context) {
	requesterID := middleware.GetUserID(c)
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}
	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	if err := h.svc.RemoveMember(tripID, requesterID, targetUserID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "member removed", nil)
}

// ---- Itinerary ----

// POST /trips/:tripId/itinerary
func (h *Handler) AddItineraryItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}
	var req CreateItineraryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.AddItineraryItem(tripID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "itinerary item added", item)
}

// GET /trips/:tripId/itinerary
func (h *Handler) GetItinerary(c *gin.Context) {
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}
	items, err := h.svc.GetItinerary(tripID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "itinerary fetched", items)
}

// PATCH /trips/:tripId/itinerary/:itemId
func (h *Handler) UpdateItineraryItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}
	var req UpdateItineraryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.UpdateItineraryItem(itemID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "itinerary item updated", item)
}

// DELETE /trips/:tripId/itinerary/:itemId
func (h *Handler) DeleteItineraryItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}
	if err := h.svc.DeleteItineraryItem(itemID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "itinerary item deleted", nil)
}

// ---- Expenses ----

// POST /trips/:tripId/expenses
func (h *Handler) AddExpense(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}
	var req CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	expense, err := h.svc.AddExpense(tripID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "expense logged", expense)
}

// GET /trips/:tripId/expenses
func (h *Handler) GetExpenses(c *gin.Context) {
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}
	expenses, err := h.svc.GetExpenses(tripID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "expenses fetched", expenses)
}

// PATCH /trips/:tripId/expenses/:expenseId
func (h *Handler) UpdateExpense(c *gin.Context) {
	userID := middleware.GetUserID(c)
	expenseID, err := uuid.Parse(c.Param("expenseId"))
	if err != nil {
		response.BadRequest(c, "invalid expense id")
		return
	}
	var req UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	expense, err := h.svc.UpdateExpense(expenseID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "expense updated", expense)
}

// DELETE /trips/:tripId/expenses/:expenseId
func (h *Handler) DeleteExpense(c *gin.Context) {
	userID := middleware.GetUserID(c)
	expenseID, err := uuid.Parse(c.Param("expenseId"))
	if err != nil {
		response.BadRequest(c, "invalid expense id")
		return
	}
	if err := h.svc.DeleteExpense(expenseID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "expense deleted", nil)
}
