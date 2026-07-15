package handlers

import (
	"hotel_lobby/internal/middleware"
	"hotel_lobby/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CustomerHandler struct {
	customerService *services.CustomerService
}

func NewCustomerHandler(cs *services.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: cs}
}

func (h *CustomerHandler) GetProfile(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals(middleware.KeyUserID).(string)
	if !ok {
		return middleware.Respond(c, fiber.StatusUnauthorized, middleware.Fail("unauthorized", "unauthorized"))
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return middleware.Respond(c, fiber.StatusUnauthorized, middleware.Fail("invalid_user", "invalid user"))
	}
	customer, err := h.customerService.GetProfile(c.Context(), userID)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("customer_not_found", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(fiber.Map{
		"id":        customer.ID.String(),
		"full_name": customer.FullName,
		"email":     customer.Email,
		"phone":     customer.Phone,
	}))
}

func (h *CustomerHandler) UpdateProfile(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals(middleware.KeyUserID).(string)
	if !ok {
		return middleware.Respond(c, fiber.StatusUnauthorized, middleware.Fail("unauthorized", "unauthorized"))
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return middleware.Respond(c, fiber.StatusUnauthorized, middleware.Fail("invalid_user", "invalid user"))
	}
	var req struct {
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
	}
	if err := c.BodyParser(&req); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_request", "invalid request body"))
	}
	if _, err := h.customerService.UpdateProfile(c.Context(), userID, req.FullName, req.Phone); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("update_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(fiber.Map{"message": "profile updated"}))
}

func (h *CustomerHandler) ListReservations(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals(middleware.KeyUserID).(string)
	if !ok {
		return middleware.Respond(c, fiber.StatusUnauthorized, middleware.Fail("unauthorized", "unauthorized"))
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return middleware.Respond(c, fiber.StatusUnauthorized, middleware.Fail("invalid_user", "invalid user"))
	}
	reservations, err := h.customerService.ListReservations(c.Context(), userID)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("fetch_failed", err.Error()))
	}
	items := make([]fiber.Map, 0, len(reservations))
	for _, r := range reservations {
		items = append(items, fiber.Map{
			"id":              r.ID.String(),
			"reference_code":  r.ReferenceCode,
			"total_amount":    r.TotalAmount,
			"status":          r.Status,
			"created_at":      r.CreatedAt,
		})
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(items))
}

func (h *CustomerHandler) GetReservation(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals(middleware.KeyUserID).(string)
	if !ok {
		return middleware.Respond(c, fiber.StatusUnauthorized, middleware.Fail("unauthorized", "unauthorized"))
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return middleware.Respond(c, fiber.StatusUnauthorized, middleware.Fail("invalid_user", "invalid user"))
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid reservation id"))
	}
	reservation, err := h.customerService.GetReservation(c.Context(), userID, id)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("reservation_not_found", err.Error()))
	}
	bookings, _ := h.customerService.FindBookingsByReservation(c.Context(), reservation.ID)
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(toReservationResponse(reservation, bookings)))
}
