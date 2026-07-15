package services

import (
	"context"
	"errors"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
)

var (
	ErrCustomerNotFound = errors.New("customer not found")
)

type CustomerService struct {
	customerRepo    CustomerRepoOps
	reservationRepo CustomerReservationRepo
}

type CustomerRepoOps interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Customer, error)
	FindByEmail(ctx context.Context, email string) (*models.Customer, error)
	Create(ctx context.Context, c *models.Customer) error
	Update(ctx context.Context, c *models.Customer) error
}

type CustomerReservationRepo interface {
	FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]models.Reservation, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Reservation, error)
	FindBookingsByReservationID(ctx context.Context, reservationID uuid.UUID) ([]models.Booking, error)
}

func NewCustomerService(cr CustomerRepoOps, rr CustomerReservationRepo) *CustomerService {
	return &CustomerService{
		customerRepo:    cr,
		reservationRepo: rr,
	}
}

func (s *CustomerService) GetProfile(ctx context.Context, customerID uuid.UUID) (*models.Customer, error) {
	c, err := s.customerRepo.FindByID(ctx, customerID)
	if err != nil {
		return nil, ErrCustomerNotFound
	}
	return c, nil
}

func (s *CustomerService) UpdateProfile(ctx context.Context, customerID uuid.UUID, name, phone string) (*models.Customer, error) {
	c, err := s.customerRepo.FindByID(ctx, customerID)
	if err != nil {
		return nil, ErrCustomerNotFound
	}

	if name != "" {
		c.FullName = name
	}
	if phone != "" {
		c.Phone = phone
	}

	if err := s.customerRepo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) ListReservations(ctx context.Context, customerID uuid.UUID) ([]models.Reservation, error) {
	_, err := s.customerRepo.FindByID(ctx, customerID)
	if err != nil {
		return nil, ErrCustomerNotFound
	}
	return s.reservationRepo.FindByCustomerID(ctx, customerID)
}

func (s *CustomerService) GetReservation(ctx context.Context, customerID, reservationID uuid.UUID) (*models.Reservation, error) {
	res, err := s.reservationRepo.FindByID(ctx, reservationID)
	if err != nil {
		return nil, err
	}
	if res.CustomerID == nil || *res.CustomerID != customerID {
		return nil, ErrCustomerNotFound
	}
	return res, nil
}

func (s *CustomerService) FindBookingsByReservation(ctx context.Context, reservationID uuid.UUID) ([]models.Booking, error) {
	return s.reservationRepo.FindBookingsByReservationID(ctx, reservationID)
}
