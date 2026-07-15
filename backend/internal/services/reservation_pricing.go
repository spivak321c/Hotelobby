package services

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
)

// getEffectiveRate returns the applicable rate for a room type and booking window,
// preferring override pricing over the base rate.
func (s *ReservationService) getEffectiveRate(
	ctx context.Context,
	roomTypeID uuid.UUID,
	checkIn, checkOut time.Time,
	bookingType string,
) (float64, error) {
	rt, err := s.roomTypeRepo.FindByID(ctx, roomTypeID)
	if err != nil {
		return 0, err
	}

	overrides, err := s.pricingRepo.FindByRoomTypeID(ctx, roomTypeID)
	if err != nil {
		return 0, err
	}

	for _, p := range overrides {
		overlaps := !checkOut.Before(p.EffectiveRange.Lower) && !checkIn.After(p.EffectiveRange.Upper)
		if p.RateType == bookingType && overlaps {
			return p.Rate, nil
		}
	}

	if bookingType == "hourly" {
		return rt.BaseRateHourly, nil
	}
	return rt.BaseRateDaily, nil
}

// calcBookingAmount computes the total charge for a single booking leg.
func calcBookingAmount(rate float64, checkIn, checkOut time.Time, bookingType string) float64 {
	hours := checkOut.Sub(checkIn).Hours()
	if bookingType == "hourly" {
		return rate * hours
	}
	nights := math.Ceil(hours / 24)
	return rate * nights
}
