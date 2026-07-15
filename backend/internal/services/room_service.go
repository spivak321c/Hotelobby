package services

import (
	"context"
	"errors"
	"math"
	"time"

	"hotel_lobby/internal/models"
	"hotel_lobby/internal/repositories"

	"github.com/google/uuid"
)

var (
	ErrRoomNotFound     = errors.New("room not found")
	ErrRoomTypeNotFound = errors.New("room type not found")
)

type RoomService struct {
	roomTypeRepo  repositories.RoomTypeRepository
	roomRepo      repositories.RoomRepository
	pricingRepo   repositories.RoomPricingRepository
	inventoryRepo repositories.RoomTypeInventoryRepository
	imageRepo     repositories.RoomImageRepository
}

func NewRoomService(
	rtRepo repositories.RoomTypeRepository,
	rRepo repositories.RoomRepository,
	pRepo repositories.RoomPricingRepository,
	iRepo repositories.RoomTypeInventoryRepository,
	imgRepo repositories.RoomImageRepository,
) *RoomService {
	return &RoomService{
		roomTypeRepo:  rtRepo,
		roomRepo:      rRepo,
		pricingRepo:   pRepo,
		inventoryRepo: iRepo,
		imageRepo:     imgRepo,
	}
}

func (s *RoomService) ListRoomTypes(ctx context.Context) ([]models.RoomType, error) {
	return s.roomTypeRepo.FindAll(ctx)
}

func (s *RoomService) CreateRoomType(ctx context.Context, name, description string, baseHourlyRate, baseDailyRate float64, maxOccupancy int, isFeatured bool) (*models.RoomType, error) {
	rt := &models.RoomType{
		ID:             uuid.New(),
		Name:           name,
		Description:    description,
		BaseRateHourly: baseHourlyRate,
		BaseRateDaily:  baseDailyRate,
		MaxOccupancy:   maxOccupancy,
		IsFeatured:     isFeatured,
	}
	if err := s.roomTypeRepo.Create(ctx, rt); err != nil {
		return nil, err
	}
	return rt, nil
}

func (s *RoomService) GetRoomType(ctx context.Context, id uuid.UUID) (*models.RoomType, error) {
	rt, err := s.roomTypeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrRoomTypeNotFound
	}
	return rt, nil
}

func (s *RoomService) ListRooms(ctx context.Context, roomTypeID *uuid.UUID, status string) ([]models.Room, error) {
	return s.roomRepo.FindAll(ctx, roomTypeID, status)
}

func (s *RoomService) GetRoom(ctx context.Context, id uuid.UUID) (*models.Room, error) {
	room, err := s.roomRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrRoomNotFound
	}
	return room, nil
}

type RoomDetail struct {
	Room   models.Room        `json:"room"`
	Images []models.RoomImage `json:"images"`
}

func (s *RoomService) GetRoomWithImages(ctx context.Context, id uuid.UUID) (*RoomDetail, error) {
	room, err := s.roomRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrRoomNotFound
	}
	images, err := s.imageRepo.FindByRoomID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &RoomDetail{Room: *room, Images: images}, nil
}

type PricingBreakdown struct {
	BaseAmount     float64  `json:"base_amount"`
	OverrideAmount *float64 `json:"override_amount,omitempty"`
	TotalAmount    float64  `json:"total_amount"`
}

func (s *RoomService) CalculatePrice(rt *models.RoomType, checkIn, checkOut time.Time, bookingType string) (*PricingBreakdown, error) {
	duration := checkOut.Sub(checkIn)
	if duration <= 0 {
		return nil, errors.New("check-out must be after check-in")
	}

	rates, err := s.pricingRepo.FindByRoomTypeID(context.Background(), rt.ID)
	if err != nil {
		return nil, err
	}

	var override *float64
	for _, r := range rates {
		if r.RateType == bookingType && !checkOut.Before(r.EffectiveRange.Lower) && !checkIn.After(r.EffectiveRange.Upper) {
			v := r.Rate
			override = &v
			break
		}
	}

	var total float64
	if bookingType == "hourly" {
		hours := duration.Hours()
		rate := rt.BaseRateHourly
		if override != nil {
			rate = *override
		}
		total = rate * hours
	} else {
		nights := math.Ceil(duration.Hours() / 24)
		rate := rt.BaseRateDaily
		if override != nil {
			rate = *override
		}
		total = rate * nights
	}

	breakdown := &PricingBreakdown{
		BaseAmount:  rt.BaseRateDaily * math.Ceil(duration.Hours()/24),
		TotalAmount: total,
	}
	if override != nil {
		breakdown.OverrideAmount = override
	}
	return breakdown, nil
}

type AvailabilityResult struct {
	Date           time.Time `json:"date"`
	TotalRooms     int       `json:"total_rooms"`
	BookedRooms    int       `json:"booked_rooms"`
	AvailableRooms int       `json:"available_rooms"`
}

func (s *RoomService) CheckAvailability(ctx context.Context, roomTypeID uuid.UUID, checkIn, checkOut time.Time) ([]AvailabilityResult, error) {
	inventory, err := s.inventoryRepo.FindByRoomTypeAndDateRange(ctx, roomTypeID, checkIn, checkOut)
	if err != nil {
		return nil, err
	}

	results := make([]AvailabilityResult, 0, len(inventory))
	for _, inv := range inventory {
		results = append(results, AvailabilityResult{
			Date:           inv.Date,
			TotalRooms:     inv.TotalRooms,
			BookedRooms:    inv.BookedRooms,
			AvailableRooms: inv.TotalRooms - inv.BookedRooms,
		})
	}
	return results, nil
}

func (s *RoomService) UpdateRoomType(ctx context.Context, id uuid.UUID, name, description *string, baseHourlyRate, baseDailyRate *float64, maxOccupancy *int, isFeatured *bool) (*models.RoomType, error) {
	rt, err := s.roomTypeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrRoomTypeNotFound
	}
	if name != nil {
		rt.Name = *name
	}
	if description != nil {
		rt.Description = *description
	}
	if baseHourlyRate != nil {
		rt.BaseRateHourly = *baseHourlyRate
	}
	if baseDailyRate != nil {
		rt.BaseRateDaily = *baseDailyRate
	}
	if maxOccupancy != nil {
		rt.MaxOccupancy = *maxOccupancy
	}
	if isFeatured != nil {
		rt.IsFeatured = *isFeatured
	}
	if err := s.roomTypeRepo.Update(ctx, rt); err != nil {
		return nil, err
	}
	return rt, nil
}

func (s *RoomService) DeleteRoomType(ctx context.Context, id uuid.UUID) error {
	return s.roomTypeRepo.Delete(ctx, id)
}

func (s *RoomService) CreateRoom(ctx context.Context, roomTypeID uuid.UUID, roomNumber, status string) (*models.Room, error) {
	if status == "" {
		status = "available"
	}
	r := &models.Room{
		ID:         uuid.New(),
		RoomTypeID: roomTypeID,
		RoomNumber: roomNumber,
		Status:     status,
	}
	if err := s.roomRepo.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *RoomService) UpdateRoom(ctx context.Context, id uuid.UUID, roomTypeID *uuid.UUID, roomNumber, status *string) (*models.Room, error) {
	room, err := s.roomRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrRoomNotFound
	}
	if roomTypeID != nil {
		room.RoomTypeID = *roomTypeID
	}
	if roomNumber != nil {
		room.RoomNumber = *roomNumber
	}
	if status != nil {
		room.Status = *status
	}
	if err := s.roomRepo.Update(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) DeleteRoom(ctx context.Context, id uuid.UUID) error {
	return s.roomRepo.Delete(ctx, id)
}

func (s *RoomService) CountActiveBookings(ctx context.Context, roomID uuid.UUID) (int, error) {
	return s.roomRepo.CountActiveBookings(ctx, roomID)
}

func (s *RoomService) CountRooms(ctx context.Context) (int, error) {
	return s.roomRepo.CountRooms(ctx)
}

func (s *RoomService) IsRoomAvailable(ctx context.Context, roomID uuid.UUID, checkIn, checkOut time.Time) (bool, error) {
	return s.roomRepo.IsAvailable(ctx, roomID, checkIn, checkOut)
}

func (s *RoomService) ListRoomImages(ctx context.Context, roomID uuid.UUID) ([]models.RoomImage, error) {
	return s.imageRepo.FindByRoomID(ctx, roomID)
}

func (s *RoomService) ListRoomPricing(ctx context.Context, roomTypeID *uuid.UUID) ([]models.RoomPricing, error) {
	return s.pricingRepo.FindAll(ctx, roomTypeID)
}

func (s *RoomService) GetRoomPricing(ctx context.Context, id uuid.UUID) (*models.RoomPricing, error) {
	rp, err := s.pricingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return rp, nil
}

func (s *RoomService) CreateRoomPricing(ctx context.Context, rp *models.RoomPricing) error {
	return s.pricingRepo.Create(ctx, rp)
}

func (s *RoomService) UpdateRoomPricing(ctx context.Context, rp *models.RoomPricing) error {
	return s.pricingRepo.Update(ctx, rp)
}

func (s *RoomService) DeleteRoomPricing(ctx context.Context, id uuid.UUID) error {
	return s.pricingRepo.Delete(ctx, id)
}

func (s *RoomService) GetInventory(ctx context.Context, roomTypeID uuid.UUID, date time.Time) (*models.RoomTypeInventory, error) {
	return s.inventoryRepo.FindByRoomTypeAndDate(ctx, roomTypeID, date)
}

func (s *RoomService) GetInventoryRange(ctx context.Context, roomTypeID uuid.UUID, from, to time.Time) ([]models.RoomTypeInventory, error) {
	return s.inventoryRepo.FindByRoomTypeAndDateRange(ctx, roomTypeID, from, to)
}

func (s *RoomService) UpdateInventory(ctx context.Context, roomTypeID uuid.UUID, date time.Time, totalRooms, bookedRooms int) error {
	return s.inventoryRepo.SetInventory(ctx, roomTypeID, date, totalRooms, bookedRooms)
}
