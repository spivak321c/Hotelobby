package services

import (
	"context"
	"testing"
	"time"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
)

type mockRoomTypeRepo struct {
	roomTypes map[uuid.UUID]*models.RoomType
}

func (m *mockRoomTypeRepo) FindAll(ctx context.Context) ([]models.RoomType, error) {
	out := make([]models.RoomType, 0, len(m.roomTypes))
	for _, rt := range m.roomTypes {
		out = append(out, *rt)
	}
	return out, nil
}

func (m *mockRoomTypeRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.RoomType, error) {
	rt, ok := m.roomTypes[id]
	if !ok {
		return nil, ErrRoomTypeNotFound
	}
	return rt, nil
}

func (m *mockRoomTypeRepo) Create(ctx context.Context, rt *models.RoomType) error {
	rt.ID = uuid.New()
	m.roomTypes[rt.ID] = rt
	return nil
}

func (m *mockRoomTypeRepo) Update(ctx context.Context, rt *models.RoomType) error {
	m.roomTypes[rt.ID] = rt
	return nil
}

func (m *mockRoomTypeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.roomTypes, id)
	return nil
}

type mockRoomRepo struct {
	rooms map[uuid.UUID]*models.Room
}

func (m *mockRoomRepo) FindAll(ctx context.Context, roomTypeID *uuid.UUID, status string) ([]models.Room, error) {
	out := make([]models.Room, 0)
	for _, r := range m.rooms {
		if roomTypeID != nil && r.RoomTypeID != *roomTypeID {
			continue
		}
		if status != "" && r.Status != status {
			continue
		}
		out = append(out, *r)
	}
	return out, nil
}

func (m *mockRoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Room, error) {
	r, ok := m.rooms[id]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return r, nil
}

func (m *mockRoomRepo) Create(ctx context.Context, r *models.Room) error {
	r.ID = uuid.New()
	m.rooms[r.ID] = r
	return nil
}

func (m *mockRoomRepo) Update(ctx context.Context, r *models.Room) error {
	m.rooms[r.ID] = r
	return nil
}

func (m *mockRoomRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.rooms, id)
	return nil
}

func (m *mockRoomRepo) CountActiveBookings(ctx context.Context, roomID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockRoomRepo) IsAvailable(ctx context.Context, roomID uuid.UUID, checkIn, checkOut time.Time) (bool, error) {
	return true, nil
}

func (m *mockRoomRepo) CountRooms(ctx context.Context) (int, error) {
	return 29, nil
}

type mockPricingRepo struct {
	prices []models.RoomPricing
}

func (m *mockPricingRepo) FindAll(ctx context.Context, roomTypeID *uuid.UUID) ([]models.RoomPricing, error) {
	if roomTypeID == nil {
		return m.prices, nil
	}
	var out []models.RoomPricing
	for _, p := range m.prices {
		if p.RoomTypeID == *roomTypeID {
			out = append(out, p)
		}
	}
	return out, nil
}
func (m *mockPricingRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.RoomPricing, error) {
	for _, p := range m.prices {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, ErrRoomTypeNotFound
}
func (m *mockPricingRepo) FindByRoomTypeID(ctx context.Context, roomTypeID uuid.UUID) ([]models.RoomPricing, error) {
	return m.prices, nil
}

func (m *mockPricingRepo) Create(ctx context.Context, rp *models.RoomPricing) error { return nil }
func (m *mockPricingRepo) Update(ctx context.Context, rp *models.RoomPricing) error { return nil }
func (m *mockPricingRepo) Delete(ctx context.Context, id uuid.UUID) error          { return nil }

type mockInventoryRepo struct {
	records []models.RoomTypeInventory
}

func (m *mockInventoryRepo) FindByRoomTypeAndDate(ctx context.Context, roomTypeID uuid.UUID, date time.Time) (*models.RoomTypeInventory, error) {
	for _, inv := range m.records {
		if inv.RoomTypeID == roomTypeID && inv.Date.Equal(date) {
			return &inv, nil
		}
	}
	return nil, nil
}

func (m *mockInventoryRepo) FindByRoomTypeAndDateRange(ctx context.Context, roomTypeID uuid.UUID, from, to time.Time) ([]models.RoomTypeInventory, error) {
	out := make([]models.RoomTypeInventory, 0)
	for _, inv := range m.records {
		if inv.RoomTypeID == roomTypeID && !inv.Date.Before(from) && !inv.Date.After(to) {
			out = append(out, inv)
		}
	}
	return out, nil
}

func (m *mockInventoryRepo) IncrementBooked(ctx context.Context, roomTypeID uuid.UUID, date time.Time) error { return nil }
func (m *mockInventoryRepo) DecrementBooked(ctx context.Context, roomTypeID uuid.UUID, date time.Time) error { return nil }
func (m *mockInventoryRepo) SetInventory(ctx context.Context, roomTypeID uuid.UUID, date time.Time, totalRooms, bookedRooms int) error {
	return nil
}

type mockImageRepo struct {
	images map[uuid.UUID][]models.RoomImage
}

func (m *mockImageRepo) FindByRoomID(ctx context.Context, roomID uuid.UUID) ([]models.RoomImage, error) {
	return m.images[roomID], nil
}

func (m *mockImageRepo) Create(ctx context.Context, img *models.RoomImage) error    { return nil }
func (m *mockImageRepo) Delete(ctx context.Context, id uuid.UUID) error             { return nil }
func (m *mockImageRepo) SetPrimary(ctx context.Context, id uuid.UUID) error         { return nil }
func (m *mockImageRepo) Reorder(ctx context.Context, roomID uuid.UUID, ids []uuid.UUID) error { return nil }

func TestRoomService_ListRoomTypes(t *testing.T) {
	id := uuid.New()
	rtRepo := &mockRoomTypeRepo{roomTypes: map[uuid.UUID]*models.RoomType{
		id: {ID: id, Name: "Deluxe", BaseRateHourly: 50, BaseRateDaily: 200},
	}}
	svc := NewRoomService(rtRepo, nil, nil, nil, nil)

	types, err := svc.ListRoomTypes(context.Background())
	if err != nil {
		t.Fatalf("ListRoomTypes: %v", err)
	}
	if len(types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(types))
	}
	if types[0].Name != "Deluxe" {
		t.Errorf("expected Deluxe, got %s", types[0].Name)
	}
}

func TestRoomService_GetRoomType_notFound(t *testing.T) {
	rtRepo := &mockRoomTypeRepo{roomTypes: map[uuid.UUID]*models.RoomType{}}
	svc := NewRoomService(rtRepo, nil, nil, nil, nil)

	_, err := svc.GetRoomType(context.Background(), uuid.New())
	if err != ErrRoomTypeNotFound {
		t.Fatalf("expected ErrRoomTypeNotFound, got %v", err)
	}
}

func TestRoomService_CalculatePrice_hourly(t *testing.T) {
	rt := &models.RoomType{BaseRateHourly: 50, BaseRateDaily: 200}
	svc := NewRoomService(nil, nil, &mockPricingRepo{}, nil, nil)

	checkIn := time.Date(2026, 7, 10, 14, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 7, 10, 18, 0, 0, 0, time.UTC)

	result, err := svc.CalculatePrice(rt, checkIn, checkOut, "hourly")
	if err != nil {
		t.Fatalf("CalculatePrice: %v", err)
	}
	expected := 50.0 * 4
	if result.TotalAmount != expected {
		t.Errorf("expected %.2f, got %.2f", expected, result.TotalAmount)
	}
	if result.OverrideAmount != nil {
		t.Error("expected no override")
	}
}

func TestRoomService_CalculatePrice_daily(t *testing.T) {
	rt := &models.RoomType{BaseRateHourly: 50, BaseRateDaily: 200}
	svc := NewRoomService(nil, nil, &mockPricingRepo{}, nil, nil)

	checkIn := time.Date(2026, 7, 10, 14, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 7, 12, 11, 0, 0, 0, time.UTC)

	result, err := svc.CalculatePrice(rt, checkIn, checkOut, "daily")
	if err != nil {
		t.Fatalf("CalculatePrice: %v", err)
	}
	expected := 200.0 * 2 // 2 nights
	if result.TotalAmount != expected {
		t.Errorf("expected %.2f, got %.2f", expected, result.TotalAmount)
	}
}

func TestRoomService_CalculatePrice_withOverride(t *testing.T) {
	overrideAmount := 350.0
	rt := &models.RoomType{BaseRateDaily: 200}
	pricingRepo := &mockPricingRepo{
		prices: []models.RoomPricing{
			{
				RoomTypeID:  rt.ID,
				RateType:    "daily",
				Rate:        overrideAmount,
				EffectiveRange: models.Daterange{
					Lower: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
					Upper: time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	svc := NewRoomService(nil, nil, pricingRepo, nil, nil)

	checkIn := time.Date(2026, 7, 10, 14, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 7, 12, 11, 0, 0, 0, time.UTC)

	result, err := svc.CalculatePrice(rt, checkIn, checkOut, "daily")
	if err != nil {
		t.Fatalf("CalculatePrice: %v", err)
	}
	expected := overrideAmount * 2
	if result.TotalAmount != expected {
		t.Errorf("expected %.2f, got %.2f", expected, result.TotalAmount)
	}
	if result.OverrideAmount == nil {
		t.Fatal("expected override amount")
	}
	if *result.OverrideAmount != overrideAmount {
		t.Errorf("expected override %.2f, got %.2f", overrideAmount, *result.OverrideAmount)
	}
}

func TestRoomService_CalculatePrice_invalidDates(t *testing.T) {
	rt := &models.RoomType{BaseRateDaily: 200}
	svc := NewRoomService(nil, nil, &mockPricingRepo{}, nil, nil)

	checkIn := time.Date(2026, 7, 12, 14, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 7, 10, 11, 0, 0, 0, time.UTC)

	_, err := svc.CalculatePrice(rt, checkIn, checkOut, "daily")
	if err == nil {
		t.Fatal("expected error for check-out before check-in")
	}
}

func TestRoomService_CheckAvailability(t *testing.T) {
	roomTypeID := uuid.New()
	today := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	invRepo := &mockInventoryRepo{
		records: []models.RoomTypeInventory{
			{RoomTypeID: roomTypeID, Date: today, TotalRooms: 10, BookedRooms: 3},
			{RoomTypeID: roomTypeID, Date: today.AddDate(0, 0, 1), TotalRooms: 10, BookedRooms: 5},
		},
	}
	svc := NewRoomService(nil, nil, nil, invRepo, nil)

	results, err := svc.CheckAvailability(context.Background(), roomTypeID, today, today.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("CheckAvailability: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].AvailableRooms != 7 {
		t.Errorf("expected 7 available, got %d", results[0].AvailableRooms)
	}
}

func TestRoomService_ListRooms_filterByStatus(t *testing.T) {
	id1, id2 := uuid.New(), uuid.New()
	roomRepo := &mockRoomRepo{
		rooms: map[uuid.UUID]*models.Room{
			id1: {ID: id1, RoomNumber: "101", Status: "available"},
			id2: {ID: id2, RoomNumber: "102", Status: "maintenance"},
		},
	}
	svc := NewRoomService(nil, roomRepo, nil, nil, nil)

	all, _ := svc.ListRooms(context.Background(), nil, "")
	if len(all) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(all))
	}

	available, _ := svc.ListRooms(context.Background(), nil, "available")
	if len(available) != 1 {
		t.Errorf("expected 1 available room, got %d", len(available))
	}
}

func TestRoomService_GetRoom_notFound(t *testing.T) {
	roomRepo := &mockRoomRepo{rooms: map[uuid.UUID]*models.Room{}}
	svc := NewRoomService(nil, roomRepo, nil, nil, nil)

	_, err := svc.GetRoom(context.Background(), uuid.New())
	if err != ErrRoomNotFound {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestRoomService_GetRoomWithImages(t *testing.T) {
	roomID := uuid.New()
	roomRepo := &mockRoomRepo{
		rooms: map[uuid.UUID]*models.Room{
			roomID: {ID: roomID, RoomNumber: "101", Status: "available"},
		},
	}
	imgRepo := &mockImageRepo{
		images: map[uuid.UUID][]models.RoomImage{
			roomID: {{ID: uuid.New(), RoomID: roomID, URL: "/img1.jpg", IsPrimary: true}},
		},
	}
	svc := NewRoomService(nil, roomRepo, nil, nil, imgRepo)

	detail, err := svc.GetRoomWithImages(context.Background(), roomID)
	if err != nil {
		t.Fatalf("GetRoomWithImages: %v", err)
	}
	if detail.Room.RoomNumber != "101" {
		t.Errorf("expected room 101, got %s", detail.Room.RoomNumber)
	}
	if len(detail.Images) != 1 {
		t.Errorf("expected 1 image, got %d", len(detail.Images))
	}
}

// ---------------------------------------------------------------------------
// CRUD tests for new RoomService methods
// ---------------------------------------------------------------------------

func TestRoomService_CreateRoomType_ok(t *testing.T) {
	rtRepo := &mockRoomTypeRepo{roomTypes: map[uuid.UUID]*models.RoomType{}}
	svc := NewRoomService(rtRepo, nil, nil, nil, nil)

	rt, err := svc.CreateRoomType(context.Background(), "Deluxe", "Luxury suite", 50, 300, 4, true)
	if err != nil {
		t.Fatalf("CreateRoomType: %v", err)
	}
	if rt.Name != "Deluxe" {
		t.Errorf("expected Deluxe, got %s", rt.Name)
	}
	if rt.BaseRateHourly != 50 {
		t.Errorf("expected hourly rate 50, got %.2f", rt.BaseRateHourly)
	}
	if rt.BaseRateDaily != 300 {
		t.Errorf("expected daily rate 300, got %.2f", rt.BaseRateDaily)
	}
	if rt.MaxOccupancy != 4 {
		t.Errorf("expected max occupancy 4, got %d", rt.MaxOccupancy)
	}
	if !rt.IsFeatured {
		t.Error("expected featured")
	}
	if _, ok := rtRepo.roomTypes[rt.ID]; !ok {
		t.Error("expected room type to be persisted")
	}
}

func TestRoomService_UpdateRoomType_ok(t *testing.T) {
	rtID := uuid.New()
	rtRepo := &mockRoomTypeRepo{roomTypes: map[uuid.UUID]*models.RoomType{
		rtID: {ID: rtID, Name: "Old", BaseRateHourly: 50, BaseRateDaily: 200},
	}}
	svc := NewRoomService(rtRepo, nil, nil, nil, nil)

	newName := "Updated"
	newRate := 100.0
	rt, err := svc.UpdateRoomType(context.Background(), rtID, &newName, nil, &newRate, nil, nil, nil)
	if err != nil {
		t.Fatalf("UpdateRoomType: %v", err)
	}
	if rt.Name != "Updated" {
		t.Errorf("expected Updated, got %s", rt.Name)
	}
	if rt.BaseRateHourly != 100 {
		t.Errorf("expected hourly 100, got %.2f", rt.BaseRateHourly)
	}
}

func TestRoomService_UpdateRoomType_notFound(t *testing.T) {
	rtRepo := &mockRoomTypeRepo{roomTypes: map[uuid.UUID]*models.RoomType{}}
	svc := NewRoomService(rtRepo, nil, nil, nil, nil)

	_, err := svc.UpdateRoomType(context.Background(), uuid.New(), nil, nil, nil, nil, nil, nil)
	if err != ErrRoomTypeNotFound {
		t.Fatalf("expected ErrRoomTypeNotFound, got %v", err)
	}
}

func TestRoomService_DeleteRoomType_ok(t *testing.T) {
	rtID := uuid.New()
	rtRepo := &mockRoomTypeRepo{roomTypes: map[uuid.UUID]*models.RoomType{
		rtID: {ID: rtID, Name: "ToDelete"},
	}}
	svc := NewRoomService(rtRepo, nil, nil, nil, nil)

	if err := svc.DeleteRoomType(context.Background(), rtID); err != nil {
		t.Fatalf("DeleteRoomType: %v", err)
	}
	if _, ok := rtRepo.roomTypes[rtID]; ok {
		t.Error("expected room type to be deleted")
	}
}

func TestRoomService_CreateRoom_ok(t *testing.T) {
	roomRepo := &mockRoomRepo{rooms: map[uuid.UUID]*models.Room{}}
	svc := NewRoomService(nil, roomRepo, nil, nil, nil)

	rtID := uuid.New()
	room, err := svc.CreateRoom(context.Background(), rtID, "101", "available")
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	if room.RoomNumber != "101" {
		t.Errorf("expected 101, got %s", room.RoomNumber)
	}
	if room.Status != "available" {
		t.Errorf("expected available, got %s", room.Status)
	}
}

func TestRoomService_CreateRoom_defaultStatus(t *testing.T) {
	roomRepo := &mockRoomRepo{rooms: map[uuid.UUID]*models.Room{}}
	svc := NewRoomService(nil, roomRepo, nil, nil, nil)

	room, err := svc.CreateRoom(context.Background(), uuid.New(), "102", "")
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	if room.Status != "available" {
		t.Errorf("expected default status available, got %s", room.Status)
	}
}

func TestRoomService_UpdateRoom_ok(t *testing.T) {
	roomID := uuid.New()
	roomRepo := &mockRoomRepo{rooms: map[uuid.UUID]*models.Room{
		roomID: {ID: roomID, RoomNumber: "101", Status: "available"},
	}}
	svc := NewRoomService(nil, roomRepo, nil, nil, nil)

	newStatus := "maintenance"
	room, err := svc.UpdateRoom(context.Background(), roomID, nil, nil, &newStatus)
	if err != nil {
		t.Fatalf("UpdateRoom: %v", err)
	}
	if room.Status != "maintenance" {
		t.Errorf("expected maintenance, got %s", room.Status)
	}
}

func TestRoomService_UpdateRoom_notFound(t *testing.T) {
	roomRepo := &mockRoomRepo{rooms: map[uuid.UUID]*models.Room{}}
	svc := NewRoomService(nil, roomRepo, nil, nil, nil)

	_, err := svc.UpdateRoom(context.Background(), uuid.New(), nil, nil, nil)
	if err != ErrRoomNotFound {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestRoomService_DeleteRoom_ok(t *testing.T) {
	roomID := uuid.New()
	roomRepo := &mockRoomRepo{rooms: map[uuid.UUID]*models.Room{
		roomID: {ID: roomID, RoomNumber: "101"},
	}}
	svc := NewRoomService(nil, roomRepo, nil, nil, nil)

	if err := svc.DeleteRoom(context.Background(), roomID); err != nil {
		t.Fatalf("DeleteRoom: %v", err)
	}
	if _, ok := roomRepo.rooms[roomID]; ok {
		t.Error("expected room to be deleted")
	}
}

func TestRoomService_ListRoomPricing_ok(t *testing.T) {
	rtID := uuid.New()
	pricingRepo := &mockPricingRepo{
		prices: []models.RoomPricing{
			{ID: uuid.New(), RoomTypeID: rtID, RateType: "daily", Rate: 300},
			{ID: uuid.New(), RoomTypeID: uuid.New(), RateType: "hourly", Rate: 50},
		},
	}
	svc := NewRoomService(nil, nil, pricingRepo, nil, nil)

	prices, err := svc.ListRoomPricing(context.Background(), &rtID)
	if err != nil {
		t.Fatalf("ListRoomPricing: %v", err)
	}
	if len(prices) != 1 {
		t.Fatalf("expected 1 pricing rule, got %d", len(prices))
	}
}

func TestRoomService_GetRoomPricing_ok(t *testing.T) {
	rpID := uuid.New()
	pricingRepo := &mockPricingRepo{
		prices: []models.RoomPricing{
			{ID: rpID, RoomTypeID: uuid.New(), RateType: "daily", Rate: 300},
		},
	}
	svc := NewRoomService(nil, nil, pricingRepo, nil, nil)

	rp, err := svc.GetRoomPricing(context.Background(), rpID)
	if err != nil {
		t.Fatalf("GetRoomPricing: %v", err)
	}
	if rp.Rate != 300 {
		t.Errorf("expected rate 300, got %.2f", rp.Rate)
	}
}

func TestRoomService_GetInventory_ok(t *testing.T) {
	rtID := uuid.New()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	invRepo := &mockInventoryRepo{
		records: []models.RoomTypeInventory{
			{RoomTypeID: rtID, Date: date, TotalRooms: 10, BookedRooms: 3},
		},
	}
	svc := NewRoomService(nil, nil, nil, invRepo, nil)

	inv, err := svc.GetInventory(context.Background(), rtID, date)
	if err != nil {
		t.Fatalf("GetInventory: %v", err)
	}
	if inv.TotalRooms != 10 {
		t.Errorf("expected 10 total rooms, got %d", inv.TotalRooms)
	}
}

func TestRoomService_UpdateInventory_ok(t *testing.T) {
	invRepo := &mockInventoryRepo{records: []models.RoomTypeInventory{}}
	svc := NewRoomService(nil, nil, nil, invRepo, nil)

	rtID := uuid.New()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	if err := svc.UpdateInventory(context.Background(), rtID, date, 15, 5); err != nil {
		t.Fatalf("UpdateInventory: %v", err)
	}
}
