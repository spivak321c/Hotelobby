package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
)

// generateOTP produces a cryptographically random 6-digit numeric OTP.
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

// RequestCancelOTP generates and stores an OTP for cancellation, then emails it.
func (s *ReservationService) RequestCancelOTP(ctx context.Context, reference, email string) error {
	reservation, err := s.reservationRepo.FindByReferenceCode(ctx, reference)
	if err != nil {
		return ErrReservationNotFound
	}
	if reservation.GuestEmail != email {
		return ErrReservationNotFound
	}
	if reservation.Status == "cancelled" {
		return ErrAlreadyCancelled
	}

	otp, err := generateOTP()
	if err != nil {
		return err
	}

	key := fmt.Sprintf("cancel_otp:%s", reference)
	if err := s.otpStore.Set(ctx, key, otp, otpTTL); err != nil {
		return err
	}

	if s.emailService != nil {
		return s.emailService.SendOTP(email, otp)
	}
	return nil
}
