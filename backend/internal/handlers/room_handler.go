package handlers

import (
	"time"

	"hotel_lobby/internal/middleware"
	"hotel_lobby/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RoomHandler struct {
	roomService *services.RoomService
}

func NewRoomHandler(rs *services.RoomService) *RoomHandler {
	return &RoomHandler{roomService: rs}
}

func (h *RoomHandler) ListRoomTypes(c *fiber.Ctx) error {
	types, err := h.roomService.ListRoomTypes(c.Context())
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("fetch_failed", err.Error()))
	}

	resp := make([]RoomTypeResponse, 0, len(types))
	for _, rt := range types {
		resp = append(resp, RoomTypeResponse{
			ID:             rt.ID.String(),
			Name:           rt.Name,
			Description:    rt.Description,
			BaseRateHourly: rt.BaseRateHourly,
			BaseRateDaily:  rt.BaseRateDaily,
			MaxOccupancy:   rt.MaxOccupancy,
			IsFeatured:     rt.IsFeatured,
		})
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(resp))
}

func (h *RoomHandler) GetRoomType(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room type id"))
	}

	rt, err := h.roomService.GetRoomType(c.Context(), id)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("room_type_not_found", err.Error()))
	}

	return middleware.Respond(c, fiber.StatusOK, middleware.OK(RoomTypeResponse{
		ID:             rt.ID.String(),
		Name:           rt.Name,
		Description:    rt.Description,
		BaseRateHourly: rt.BaseRateHourly,
		BaseRateDaily:  rt.BaseRateDaily,
		MaxOccupancy:   rt.MaxOccupancy,
		IsFeatured:     rt.IsFeatured,
	}))
}

func (h *RoomHandler) ListRooms(c *fiber.Ctx) error {
	var roomTypeID *uuid.UUID
	if rt := c.Query("room_type_id"); rt != "" {
		parsed, err := uuid.Parse(rt)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_room_type_id", "invalid room_type_id"))
		}
		roomTypeID = &parsed
	}

	rooms, err := h.roomService.ListRooms(c.Context(), roomTypeID, c.Query("status"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("fetch_failed", err.Error()))
	}

	resp := make([]RoomResponse, 0, len(rooms))
	for _, r := range rooms {
		count, _ := h.roomService.CountActiveBookings(c.Context(), r.ID)
		resp = append(resp, RoomResponse{
			ID:               r.ID.String(),
			RoomTypeID:       r.RoomTypeID.String(),
			RoomNumber:       r.RoomNumber,
			Status:           r.Status,
			UpcomingBookings: count,
		})
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(resp))
}

func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room id"))
	}

	room, err := h.roomService.GetRoomWithImages(c.Context(), id)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("room_not_found", err.Error()))
	}

	rt, err := h.roomService.GetRoomType(c.Context(), room.Room.RoomTypeID)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("room_type_not_found", "failed to load room type"))
	}

	imgResp := make([]RoomImageResponse, 0, len(room.Images))
	for _, img := range room.Images {
		imgResp = append(imgResp, RoomImageResponse{
			ID:        img.ID.String(),
			URL:       img.URL,
			IsPrimary: img.IsPrimary,
			SortOrder: img.SortOrder,
		})
	}

	bookingCount, _ := h.roomService.CountActiveBookings(c.Context(), id)

	return middleware.Respond(c, fiber.StatusOK, middleware.OK(RoomDetailResponse{
		Room: RoomResponse{
			ID:               room.Room.ID.String(),
			RoomTypeID:       room.Room.RoomTypeID.String(),
			RoomNumber:       room.Room.RoomNumber,
			Status:           room.Room.Status,
			UpcomingBookings: bookingCount,
		},
		RoomTypeName:   rt.Name,
		BaseRateDaily:  rt.BaseRateDaily,
		BaseRateHourly: rt.BaseRateHourly,
		Images:         imgResp,
	}))
}

func (h *RoomHandler) ListImages(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room id"))
	}

	images, err := h.roomService.ListRoomImages(c.Context(), id)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("fetch_failed", err.Error()))
	}

	resp := make([]RoomImageResponse, 0, len(images))
	for _, img := range images {
		resp = append(resp, RoomImageResponse{
			ID:        img.ID.String(),
			URL:       img.URL,
			IsPrimary: img.IsPrimary,
			SortOrder: img.SortOrder,
		})
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(resp))
}

func (h *RoomHandler) CheckAvailability(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room id"))
	}

	checkIn, err := time.Parse("2006-01-02", c.Query("check_in"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid check_in date (YYYY-MM-DD)"))
	}
	checkOut, err := time.Parse("2006-01-02", c.Query("check_out"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid check_out date (YYYY-MM-DD)"))
	}

	bookingType := c.Query("type", "daily")

	room, err := h.roomService.GetRoom(c.Context(), id)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("room_not_found", "room not found"))
	}

	rt, err := h.roomService.GetRoomType(c.Context(), room.RoomTypeID)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("room_type_not_found", "room type not found"))
	}

	pricing, err := h.roomService.CalculatePrice(rt, checkIn, checkOut, bookingType)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("price_calculation_failed", err.Error()))
	}

	results, err := h.roomService.CheckAvailability(c.Context(), room.RoomTypeID, checkIn, checkOut)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("availability_check_failed", err.Error()))
	}

	minAvailable := int(^uint(0) >> 1)
	dates := make([]AvailabilityDate, 0, len(results))
	for _, a := range results {
		avail := a.AvailableRooms
		if avail < minAvailable {
			minAvailable = avail
		}
		dates = append(dates, AvailabilityDate{
			Date:      a.Date.Format("2006-01-02"),
			Available: avail,
		})
	}

	if len(results) == 0 {
		minAvailable = 0
	}

	isAvailable, err := h.roomService.IsRoomAvailable(c.Context(), id, checkIn, checkOut)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("availability_check_failed", "failed to check specific room availability"))
	}

	return middleware.Respond(c, fiber.StatusOK, middleware.OK(AvailabilityResponse{
		RoomTypeID:     room.RoomTypeID.String(),
		Available:      isAvailable,
		AvailableRooms: minAvailable,
		TotalPrice:     pricing.TotalAmount,
		Dates:          dates,
	}))
}
