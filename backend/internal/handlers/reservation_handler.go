package handlers

import (
	"time"

	"hotel_lobby/internal/middleware"
	"hotel_lobby/internal/models"
	"hotel_lobby/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ReservationHandler struct {
	reservationService *services.ReservationService
}

func NewReservationHandler(rs *services.ReservationService) *ReservationHandler {
	return &ReservationHandler{reservationService: rs}
}

func (h *ReservationHandler) Create(c *fiber.Ctx) error {
	var req CreateReservationRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}

	bookings := make([]services.BookingInput, 0, len(req.Bookings))
	for _, b := range req.Bookings {
		roomID, err := uuid.Parse(b.RoomID)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_room_id", "invalid room_id in bookings"))
		}
		checkIn, err := time.Parse("2006-01-02", b.CheckIn)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid check_in date, expected YYYY-MM-DD"))
		}
		checkOut, err := time.Parse("2006-01-02", b.CheckOut)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid check_out date, expected YYYY-MM-DD"))
		}
		bookings = append(bookings, services.BookingInput{
			RoomID:            roomID,
			CheckIn:           checkIn,
			CheckOut:          checkOut,
			BookingType:       b.BookingType,
			ExpectedOccupants: b.ExpectedOccupants,
		})
	}

	input := services.CreateReservationInput{
		GuestName:     req.GuestName,
		GuestEmail:    req.GuestEmail,
		GuestPhone:    req.GuestPhone,
		Bookings:      bookings,
		PaymentMethod: req.PaymentMethod,
	}

	ik := c.Get("Idempotency-Key")
	if ik != "" {
		input.IdempotencyKey = &ik
	}

	reservation, err := h.reservationService.Create(c.Context(), input)
	if err != nil {
		code := "creation_failed"
		status := fiber.StatusInternalServerError
		switch err {
		case services.ErrMaxRoomsExceeded:
			code = "max_rooms_exceeded"
			status = fiber.StatusBadRequest
		case services.ErrInvalidBookingDates:
			code = "invalid_booking_dates"
			status = fiber.StatusBadRequest
		case services.ErrRoomNotAvailable:
			code = "room_not_available"
			status = fiber.StatusBadRequest
		case services.ErrRoomNotFound:
			code = "room_not_found"
			status = fiber.StatusNotFound
		}
		return middleware.Respond(c, status, middleware.Fail(code, err.Error()))
	}

	resBookings, _ := h.reservationService.FindAllBookingsByReservation(c.Context(), reservation.ID)
	return middleware.Respond(c, fiber.StatusCreated, middleware.Created(toReservationResponse(reservation, resBookings)))
}

func (h *ReservationHandler) Lookup(c *fiber.Ctx) error {
	reference := c.Params("reference")
	email := c.Query("email")
	if email == "" {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("missing_email", "email query parameter is required"))
	}

	reservation, err := h.reservationService.Lookup(c.Context(), reference, email)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("reservation_not_found", err.Error()))
	}

	bookings, _ := h.reservationService.FindAllBookingsByReservation(c.Context(), reservation.ID)
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(toReservationResponse(reservation, bookings)))
}

func (h *ReservationHandler) RequestCancelOTP(c *fiber.Ctx) error {
	var req RequestOTPRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}

	err := h.reservationService.RequestCancelOTP(c.Context(), c.Params("reference"), req.Email)
	if err != nil {
		code := "otp_request_failed"
		status := fiber.StatusInternalServerError
		switch err {
		case services.ErrReservationNotFound:
			code = "reservation_not_found"
			status = fiber.StatusNotFound
		case services.ErrAlreadyCancelled:
			code = "already_cancelled"
			status = fiber.StatusBadRequest
		}
		return middleware.Respond(c, status, middleware.Fail(code, err.Error()))
	}

	return middleware.Respond(c, fiber.StatusOK, middleware.OK(fiber.Map{"message": "OTP sent to your email"}))
}

func (h *ReservationHandler) Cancel(c *fiber.Ctx) error {
	var req CancelRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}

	err := h.reservationService.Cancel(c.Context(), c.Params("reference"), req.Otp, req.Reason)
	if err != nil {
		code := "cancel_failed"
		status := fiber.StatusInternalServerError
		switch err {
		case services.ErrReservationNotFound:
			code = "reservation_not_found"
			status = fiber.StatusNotFound
		case services.ErrInvalidOTP:
			code = "invalid_otp"
			status = fiber.StatusBadRequest
		case services.ErrAlreadyCancelled:
			code = "already_cancelled"
			status = fiber.StatusBadRequest
		}
		return middleware.Respond(c, status, middleware.Fail(code, err.Error()))
	}

	return middleware.Respond(c, fiber.StatusOK, middleware.OK(fiber.Map{"message": "reservation cancelled"}))
}

func toReservationResponse(r *models.Reservation, bookings []models.Booking) ReservationResponse {
	resp := ReservationResponse{
		ID:                 r.ID.String(),
		ReferenceCode:      r.ReferenceCode,
		GuestName:          r.GuestName,
		GuestEmail:         r.GuestEmail,
		GuestPhone:         r.GuestPhone,
		TotalAmount:        r.TotalAmount,
		Status:             r.Status,
		CancellationReason: r.CancellationReason,
		CreatedAt:          r.CreatedAt,
	}
	resp.Bookings = make([]BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		resp.Bookings = append(resp.Bookings, BookingResponse{
			ID:          b.ID.String(),
			RoomID:      b.RoomID.String(),
			RoomTypeID:  b.RoomTypeID.String(),
			BookingType: b.BookingType,
			StartsAt:    b.StartsAt,
			EndsAt:      b.EndsAt,
			Status:      b.Status,
			Amount:      b.Amount,
		})
	}
	return resp
}
