package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"hotel_lobby/internal/middleware"
	"hotel_lobby/internal/models"
	"hotel_lobby/internal/repositories"
	"hotel_lobby/internal/services"
	"hotel_lobby/internal/sse"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	roomService        *services.RoomService
	reservationService *services.ReservationService
	bookingService     *services.BookingService
	inventoryService   *services.InventoryService
	authService        *services.AuthService
	imageService       *services.ImageService
	customerRepo       repositories.CustomerRepository
	adminRepo          repositories.AdminRepository
	sseHub             *sse.Hub
}

func NewAdminHandler(
	roomService *services.RoomService,
	reservationService *services.ReservationService,
	bookingService *services.BookingService,
	inventoryService *services.InventoryService,
	authService *services.AuthService,
	imageService *services.ImageService,
	customerRepo repositories.CustomerRepository,
	adminRepo repositories.AdminRepository,
	sseHub *sse.Hub,
) *AdminHandler {
	return &AdminHandler{
		roomService:        roomService,
		reservationService: reservationService,
		bookingService:     bookingService,
		inventoryService:   inventoryService,
		authService:        authService,
		imageService:       imageService,
		customerRepo:       customerRepo,
		adminRepo:          adminRepo,
		sseHub:             sseHub,
	}
}

// Room Types

func (h *AdminHandler) ListRoomTypes(c *fiber.Ctx) error {
	types, err := h.roomService.ListRoomTypes(c.Context())
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("fetch_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(types))
}

func (h *AdminHandler) CreateRoomType(c *fiber.Ctx) error {
	var req struct {
		Name          string  `json:"name" validate:"required"`
		Description   string  `json:"description"`
		BaseHourlyRate float64 `json:"base_rate_hourly" validate:"required,min=0"`
		BaseDailyRate  float64 `json:"base_rate_daily" validate:"required,min=0"`
		MaxOccupancy   int     `json:"max_occupancy" validate:"min=1"`
		IsFeatured     bool    `json:"is_featured"`
	}
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	rt, err := h.roomService.CreateRoomType(c.Context(), req.Name, req.Description, req.BaseHourlyRate, req.BaseDailyRate, req.MaxOccupancy, req.IsFeatured)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("create_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusCreated, middleware.Created(rt))
}

func (h *AdminHandler) UpdateRoomType(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room type id"))
	}

	var req struct {
		Name            *string  `json:"name"`
		Description     *string  `json:"description"`
		BaseHourlyRate  *float64 `json:"base_rate_hourly"`
		BaseDailyRate   *float64 `json:"base_rate_daily"`
		MaxOccupancy    *int     `json:"max_occupancy"`
		IsFeatured      *bool    `json:"is_featured"`
	}
	if err := c.BodyParser(&req); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_body", "invalid request body"))
	}

	rt, err := h.roomService.UpdateRoomType(c.Context(), id, req.Name, req.Description, req.BaseHourlyRate, req.BaseDailyRate, req.MaxOccupancy, req.IsFeatured)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("update_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(rt))
}

func (h *AdminHandler) DeleteRoomType(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room type id"))
	}

	if err := h.roomService.DeleteRoomType(c.Context(), id); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("delete_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(nil))
}

// Rooms

func (h *AdminHandler) ListRooms(c *fiber.Ctx) error {
	var rtID *uuid.UUID
	if rtStr := c.Query("room_type_id"); rtStr != "" {
		parsed, err := uuid.Parse(rtStr)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_room_type_id", "invalid room_type_id"))
		}
		rtID = &parsed
	}
	status := c.Query("status")

	rooms, err := h.roomService.ListRooms(c.Context(), rtID, status)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("fetch_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(rooms))
}

func (h *AdminHandler) CreateRoom(c *fiber.Ctx) error {
	var req struct {
		RoomTypeID string `json:"room_type_id"`
		RoomNumber string `json:"room_number"`
		Status     string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_body", "invalid request body"))
	}
	if req.RoomTypeID == "" {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("missing_field", "room_type_id is required"))
	}
	if req.RoomNumber == "" {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("missing_field", "room_number is required"))
	}

	roomTypeID, err := uuid.Parse(req.RoomTypeID)
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room_type_id"))
	}

	room, err := h.roomService.CreateRoom(c.Context(), roomTypeID, req.RoomNumber, req.Status)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("create_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusCreated, middleware.Created(room))
}

func (h *AdminHandler) UpdateRoom(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room id"))
	}

	var req struct {
		RoomTypeID *string `json:"room_type_id"`
		RoomNumber *string `json:"room_number"`
		Status     *string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_body", "invalid request body"))
	}

	var roomTypeID *uuid.UUID
	if req.RoomTypeID != nil {
		parsed, err := uuid.Parse(*req.RoomTypeID)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room_type_id"))
		}
		roomTypeID = &parsed
	}

	room, err := h.roomService.UpdateRoom(c.Context(), id, roomTypeID, req.RoomNumber, req.Status)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("update_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(room))
}

func (h *AdminHandler) DeleteRoom(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room id"))
	}

	if err := h.roomService.DeleteRoom(c.Context(), id); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("delete_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(nil))
}

// Images

func (h *AdminHandler) UploadImage(c *fiber.Ctx) error {
	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room id"))
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("missing_file", "image file is required"))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("file_error", "could not open file"))
	}
	defer file.Close()

	isPrimary := c.FormValue("is_primary") == "true"
	sortOrder := 0
	if v := c.FormValue("sort_order"); v != "" {
		sortOrder, _ = strconv.Atoi(v)
	}

	img, err := h.imageService.Upload(c.Context(), roomID, file, isPrimary, sortOrder)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("upload_failed", err.Error()))
	}

	return middleware.Respond(c, fiber.StatusCreated, middleware.Created(img))
}

func (h *AdminHandler) DeleteImage(c *fiber.Ctx) error {
	imageID, err := uuid.Parse(c.Params("image_id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid image id"))
	}
	if err := h.imageService.Delete(c.Context(), imageID); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("delete_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(nil))
}

func (h *AdminHandler) ReorderImages(c *fiber.Ctx) error {
	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room id"))
	}
	var req struct {
		ImageIDs []string `json:"image_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_body", "invalid request body"))
	}
	ids := make([]uuid.UUID, 0, len(req.ImageIDs))
	for _, s := range req.ImageIDs {
		parsed, err := uuid.Parse(s)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid image_id in list"))
		}
		ids = append(ids, parsed)
	}
	if err := h.imageService.Reorder(c.Context(), roomID, ids); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("reorder_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(nil))
}

// Pricing

func (h *AdminHandler) ListRoomPricing(c *fiber.Ctx) error {
	var roomTypeID *uuid.UUID
	if rt := c.Query("room_type_id"); rt != "" {
		parsed, err := uuid.Parse(rt)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room_type_id"))
		}
		roomTypeID = &parsed
	}
	pricing, err := h.roomService.ListRoomPricing(c.Context(), roomTypeID)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("fetch_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(pricing))
}

func (h *AdminHandler) CreateRoomPricing(c *fiber.Ctx) error {
	var req struct {
		RoomTypeID    string  `json:"room_type_id"`
		RateType      string  `json:"rate_type"`
		Rate          float64 `json:"rate"`
		EffectiveFrom string  `json:"effective_from"`
		EffectiveTo   string  `json:"effective_to"`
	}
	if err := c.BodyParser(&req); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_body", "invalid request body"))
	}
	if req.RoomTypeID == "" || req.RateType == "" {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("missing_field", "room_type_id and rate_type are required"))
	}
	rtID, err := uuid.Parse(req.RoomTypeID)
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room_type_id"))
	}
	from, err := time.Parse("2006-01-02", req.EffectiveFrom)
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid effective_from"))
	}
	to, err := time.Parse("2006-01-02", req.EffectiveTo)
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid effective_to"))
	}
	rp := &models.RoomPricing{
		ID:             uuid.New(),
		RoomTypeID:     rtID,
		RateType:       req.RateType,
		Rate:           req.Rate,
		EffectiveRange: models.Daterange{Lower: from, Upper: to},
	}
	if err := h.roomService.CreateRoomPricing(c.Context(), rp); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("create_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusCreated, middleware.Created(rp))
}

func (h *AdminHandler) UpdateRoomPricing(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid pricing id"))
	}
	rp, err := h.roomService.GetRoomPricing(c.Context(), id)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("not_found", "pricing rule not found"))
	}
	var req struct {
		RateType      *string  `json:"rate_type"`
		Rate          *float64 `json:"rate"`
		EffectiveFrom *string  `json:"effective_from"`
		EffectiveTo   *string  `json:"effective_to"`
	}
	if err := c.BodyParser(&req); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_body", "invalid request body"))
	}
	if req.RateType != nil {
		rp.RateType = *req.RateType
	}
	if req.Rate != nil {
		rp.Rate = *req.Rate
	}
	if req.EffectiveFrom != nil {
		from, err := time.Parse("2006-01-02", *req.EffectiveFrom)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid effective_from"))
		}
		rp.EffectiveRange.Lower = from
	}
	if req.EffectiveTo != nil {
		to, err := time.Parse("2006-01-02", *req.EffectiveTo)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid effective_to"))
		}
		rp.EffectiveRange.Upper = to
	}
	if err := h.roomService.UpdateRoomPricing(c.Context(), rp); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("update_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(rp))
}

func (h *AdminHandler) DeleteRoomPricing(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid pricing id"))
	}
	if err := h.roomService.DeleteRoomPricing(c.Context(), id); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("delete_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(nil))
}

// Inventory

func (h *AdminHandler) GetInventory(c *fiber.Ctx) error {
	dateStr := c.Query("date")
	if dateStr == "" {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("missing_param", "date is required"))
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid date format"))
	}
	rtIDStr := c.Query("room_type_id")
	if rtIDStr != "" {
		rtID, err := uuid.Parse(rtIDStr)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room_type_id"))
		}
		inv, err := h.roomService.GetInventory(c.Context(), rtID, date)
		if err != nil {
			return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("not_found", err.Error()))
		}
		return middleware.Respond(c, fiber.StatusOK, middleware.OK([]models.RoomTypeInventory{*inv}))
	}
	types, err := h.roomService.ListRoomTypes(c.Context())
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("fetch_failed", err.Error()))
	}
	results := make([]models.RoomTypeInventory, 0, len(types))
	for _, rt := range types {
		inv, err := h.roomService.GetInventory(c.Context(), rt.ID, date)
		if err != nil {
			continue
		}
		results = append(results, *inv)
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(results))
}

func (h *AdminHandler) UpdateInventory(c *fiber.Ctx) error {
	var req struct {
		RoomTypeID string `json:"room_type_id"`
		Date       string `json:"date"`
		TotalRooms int    `json:"total_rooms"`
		BookedRooms int   `json:"booked_rooms"`
	}
	if err := c.BodyParser(&req); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_body", "invalid request body"))
	}
	rtID, err := uuid.Parse(req.RoomTypeID)
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid room_type_id"))
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid date format"))
	}
	if err := h.roomService.UpdateInventory(c.Context(), rtID, date, req.TotalRooms, req.BookedRooms); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("update_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(nil))
}

// Reservations

func (h *AdminHandler) ListReservations(c *fiber.Ctx) error {
	status := c.Query("status")
	var from, to time.Time
	var err error
	if f := c.Query("from"); f != "" {
		from, err = time.Parse("2006-01-02", f)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid from date"))
		}
	}
	if t := c.Query("to"); t != "" {
		to, err = time.Parse("2006-01-02", t)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid to date"))
		}
	}
	reservations, err := h.reservationService.FindAll(c.Context(), status, from, to)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("fetch_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(reservations))
}

func (h *AdminHandler) GetReservation(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid reservation id"))
	}
	res, err := h.reservationService.FindByID(c.Context(), id)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("not_found", "reservation not found"))
	}
	bookings, _ := h.reservationService.FindAllBookingsByReservation(c.Context(), res.ID)
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(toReservationResponse(res, bookings)))
}

func (h *AdminHandler) UpdateReservationStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid reservation id"))
	}
	var req struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_body", "invalid request body"))
	}
	validStatuses := map[string]bool{"pending": true, "confirmed": true, "checked_in": true, "checked_out": true, "cancelled": true, "refunded": true}
	if !validStatuses[req.Status] {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_status", "invalid status value"))
	}
	if err := h.reservationService.UpdateStatus(c.Context(), id, req.Status, req.Reason); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("update_failed", err.Error()))
	}

	updatedReservation, err := h.reservationService.FindByID(c.Context(), id)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("not_found", "reservation not found after update"))
	}
	bookings, _ := h.reservationService.FindAllBookingsByReservation(c.Context(), id)
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(toReservationResponse(updatedReservation, bookings)))
}

// Customers

func (h *AdminHandler) ListCustomers(c *fiber.Ctx) error {
	customers, err := h.customerRepo.FindAll(c.Context())
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("fetch_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(customers))
}

func (h *AdminHandler) GetCustomer(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid customer id"))
	}
	customer, err := h.customerRepo.FindByID(c.Context(), id)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("not_found", "customer not found"))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(customer))
}

// Admins

func (h *AdminHandler) ListAdmins(c *fiber.Ctx) error {
	admins, err := h.adminRepo.FindAll(c.Context())
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("fetch_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(admins))
}

func (h *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	var req struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.BodyParser(&req); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_body", "invalid request body"))
	}
	if req.FullName == "" || req.Email == "" || req.Password == "" {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("missing_field", "full_name, email, and password are required"))
	}
	if req.Role == "" {
		req.Role = "front_desk"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("hash_failed", "failed to hash password"))
	}
	admin := &models.Admin{
		ID:           uuid.New(),
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         req.Role,
		IsActive:     true,
	}
	if err := h.adminRepo.Create(c.Context(), admin); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("create_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusCreated, middleware.Created(admin))
}

func (h *AdminHandler) UpdateAdmin(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid admin id"))
	}
	admin, err := h.adminRepo.FindByID(c.Context(), id)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("not_found", "admin not found"))
	}
	var req struct {
		FullName *string `json:"full_name"`
		Email    *string `json:"email"`
		Role     *string `json:"role"`
		IsActive *bool   `json:"is_active"`
		Password *string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_body", "invalid request body"))
	}
	if req.FullName != nil {
		admin.FullName = *req.FullName
	}
	if req.Email != nil {
		if *req.Email != admin.Email {
			existing, err := h.adminRepo.FindByEmail(c.Context(), *req.Email)
			if err == nil && existing.ID != id {
				return middleware.Respond(c, fiber.StatusConflict, middleware.Fail("email_taken", "email is already in use by another admin"))
			}
		}
		admin.Email = *req.Email
	}
	if req.Role != nil {
		admin.Role = *req.Role
	}
	if req.IsActive != nil {
		admin.IsActive = *req.IsActive
	}
	if req.Password != nil && *req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("hash_failed", "failed to hash password"))
		}
		admin.PasswordHash = string(hash)
	}
	if err := h.adminRepo.Update(c.Context(), admin); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("update_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(admin))
}

func (h *AdminHandler) DeleteAdmin(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_id", "invalid admin id"))
	}
	if err := h.adminRepo.Delete(c.Context(), id); err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("delete_failed", err.Error()))
	}
	return middleware.Respond(c, fiber.StatusOK, middleware.OK(nil))
}

// Reports

func (h *AdminHandler) BookingReport(c *fiber.Ctx) error {
	ctx := c.Context()

	fromStr := c.Query("from")
	toStr := c.Query("to")

	var fromTime, toTime time.Time
	var err error
	if toStr == "" {
		toTime = time.Now()
		toStr = toTime.Format("2006-01-02")
	} else {
		toTime, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid 'to' date, expected YYYY-MM-DD"))
		}
	}
	if fromStr == "" {
		fromTime = toTime.AddDate(0, 0, -30)
		fromStr = fromTime.Format("2006-01-02")
	} else {
		fromTime, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid 'from' date, expected YYYY-MM-DD"))
		}
	}

	reservations, err := h.reservationService.FindAll(ctx, "", fromTime, toTime)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("report_failed", err.Error()))
	}

	totalBookings := len(reservations)
	var totalRevenue float64
	byStatus := map[string]int{}
	for _, r := range reservations {
		byStatus[r.Status]++
		if r.Status == "confirmed" || r.Status == "checked_out" {
			totalRevenue += r.TotalAmount
		}
	}

	return middleware.Respond(c, fiber.StatusOK, middleware.OK(fiber.Map{
		"from":           fromStr,
		"to":             toStr,
		"total_bookings": totalBookings,
		"total_revenue":  totalRevenue,
		"by_status":      byStatus,
	}))
}

func (h *AdminHandler) OccupancyReport(c *fiber.Ctx) error {
	ctx := c.Context()

	fromStr := c.Query("from")
	toStr := c.Query("to")

	var fromTime, toTime time.Time
	var err error
	if toStr == "" {
		toTime = time.Now()
		toStr = toTime.Format("2006-01-02")
	} else {
		toTime, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid 'to' date, expected YYYY-MM-DD"))
		}
	}
	if fromStr == "" {
		fromTime = toTime.AddDate(0, 0, -30)
		fromStr = fromTime.Format("2006-01-02")
	} else {
		fromTime, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid 'from' date, expected YYYY-MM-DD"))
		}
	}

	reservations, err := h.reservationService.FindAll(ctx, "", fromTime, toTime)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("report_failed", err.Error()))
	}

	totalRooms, err := h.roomService.CountRooms(ctx)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("report_failed", err.Error()))
	}
	occupied := 0
	for _, r := range reservations {
		if r.Status == "confirmed" || r.Status == "checked_in" {
			occupied++
		}
	}
	if occupied > totalRooms {
		occupied = totalRooms
	}
	rate := float64(occupied) / float64(totalRooms) * 100

	return middleware.Respond(c, fiber.StatusOK, middleware.OK(fiber.Map{
		"from":            fromStr,
		"to":              toStr,
		"total_rooms":     totalRooms,
		"occupied_rooms":  occupied,
		"occupancy_rate":  rate,
	}))
}

func (h *AdminHandler) RevenueReport(c *fiber.Ctx) error {
	ctx := c.Context()

	fromStr := c.Query("from")
	toStr := c.Query("to")

	var fromTime, toTime time.Time
	var err error
	if toStr == "" {
		toTime = time.Now()
		toStr = toTime.Format("2006-01-02")
	} else {
		toTime, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid 'to' date, expected YYYY-MM-DD"))
		}
	}
	if fromStr == "" {
		fromTime = toTime.AddDate(0, 0, -30)
		fromStr = fromTime.Format("2006-01-02")
	} else {
		fromTime, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid 'from' date, expected YYYY-MM-DD"))
		}
	}

	reservations, err := h.reservationService.FindAll(ctx, "", fromTime, toTime)
	if err != nil {
		return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("report_failed", err.Error()))
	}

	byStatus := map[string]float64{}
	for _, r := range reservations {
		byStatus[r.Status] += r.TotalAmount
	}
	totalRevenue := byStatus["confirmed"] + byStatus["checked_out"]
	cancelledRevenue := byStatus["cancelled"]

	return middleware.Respond(c, fiber.StatusOK, middleware.OK(fiber.Map{
		"from":              fromStr,
		"to":                toStr,
		"total_revenue":     totalRevenue,
		"by_status":         byStatus,
		"cancelled_revenue": cancelledRevenue,
	}))
}

// SSE

func (h *AdminHandler) SSEEvents(c *fiber.Ctx) error {
	ch := make(chan sse.Event, 64)

	// Try JWT from query param first (used by EventSource), fall back to header
	tokenStr := c.Query("token")
	if tokenStr == "" {
		auth := c.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr = auth[7:]
		}
	}

	var userID, sessionID string
	if tokenStr != "" {
		if uid, _, err := h.authService.ValidateToken(tokenStr); err == nil {
			userID = uid
		}
	}

	if userID != "" {
		h.sseHub.SubscribeUser(userID, ch)
	} else {
		sessionID = c.Get("X-Session-ID")
		if sessionID == "" {
			sessionID = "anon"
		}
		h.sseHub.SubscribeGuest(sessionID, ch)
	}

	// Capture origin before hijacking (ctx is unsafe inside the goroutine).
	origin := string(c.Context().Request.Header.Peek("Origin"))
	if origin == "" {
		origin = "*"
	}

	// Hijack the raw connection — bypass fasthttp's normal response writing
	c.Context().HijackSetNoResponse(true)
	c.Context().Hijack(func(conn net.Conn) {
		defer conn.Close()
		if userID != "" {
			defer h.sseHub.UnsubscribeUser(userID, ch)
		} else {
			defer h.sseHub.UnsubscribeGuest(sessionID, ch)
		}

		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\nCache-Control: no-cache\r\nConnection: keep-alive\r\nAccess-Control-Allow-Origin: %s\r\nAccess-Control-Allow-Credentials: true\r\nX-Accel-Buffering: no\r\n\r\n", origin)))

		// Initial flush — let the browser know the connection is alive.
		conn.Write([]byte(":ok\n\n"))

		heartbeat := time.NewTicker(30 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case <-heartbeat.C:
				if _, err := conn.Write([]byte(":heartbeat\n\n")); err != nil {
					return
				}
			case event, ok := <-ch:
				if !ok {
					return
				}
				data, _ := json.Marshal(event.Payload)
				line := fmt.Sprintf("event: %s\ndata: %s\n\n", event.Type, data)
				if _, err := conn.Write([]byte(line)); err != nil {
					return
				}
			}
		}
	})

	return nil
}

func (h *AdminHandler) CreateWalkIn(c *fiber.Ctx) error {
	var req struct {
		RoomID            string `json:"room_id"            validate:"required,uuid"`
		CheckIn           string `json:"check_in"           validate:"required"`
		CheckOut          string `json:"check_out"          validate:"required"`
		BookingType       string `json:"booking_type"       validate:"required,oneof=daily hourly"`
		ExpectedOccupants int    `json:"expected_occupants" validate:"min=1"`
		Amount            float64 `json:"amount"            validate:"min=0"`
		GuestName         string `json:"guest_name"         validate:"required"`
		GuestEmail        string `json:"guest_email"        validate:"required,email"`
		GuestPhone        string `json:"guest_phone"        validate:"required"`
	}
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}

	roomID, err := uuid.Parse(req.RoomID)
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_room_id", "invalid room_id"))
	}
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid check_in date, expected YYYY-MM-DD"))
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_date", "invalid check_out date, expected YYYY-MM-DD"))
	}

	reservation, err := h.bookingService.CreateWalkIn(c.Context(), services.WalkInBookingInput{
		RoomID:            roomID,
		CheckIn:           checkIn,
		CheckOut:          checkOut,
		BookingType:       req.BookingType,
		ExpectedOccupants: req.ExpectedOccupants,
		Amount:            req.Amount,
		GuestName:         req.GuestName,
		GuestEmail:        req.GuestEmail,
		GuestPhone:        req.GuestPhone,
	})
	if err != nil {
		code := "creation_failed"
		status := fiber.StatusInternalServerError
		switch err {
		case services.ErrRoomNotFound:
			code = "room_not_found"
			status = fiber.StatusNotFound
		case services.ErrRoomNotAvailable:
			code = "room_not_available"
			status = fiber.StatusBadRequest
		}
		return middleware.Respond(c, status, middleware.Fail(code, err.Error()))
	}

	bookings, _ := h.reservationService.FindAllBookingsByReservation(c.Context(), reservation.ID)
	return middleware.Respond(c, fiber.StatusCreated, middleware.Created(toReservationResponse(reservation, bookings)))
}
