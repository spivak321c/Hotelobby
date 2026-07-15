package services

import (
	"fmt"
	"net/smtp"
	"strings"
)

type EmailService struct {
	host     string
	port     int
	user     string
	pass     string
	fromAddr string
	fromName string
	appURL   string
}

type EmailConfig struct {
	SMTPServer string
	SMTPPort   int
	SMTPUser   string
	SMTPPass   string
	EmailFrom  string // "Hotel Lobby <noreply@hotellobby.com>"
	AppURL     string
}

func parseFrom(from string) (name, addr string) {
	if idx := strings.Index(from, "<"); idx != -1 {
		name = strings.TrimSpace(from[:idx])
		addr = from[idx+1 : len(from)-1]
		return
	}
	return "", from
}

func NewEmailService(cfg EmailConfig) *EmailService {
	name, addr := parseFrom(cfg.EmailFrom)
	return &EmailService{
		host:     cfg.SMTPServer,
		port:     cfg.SMTPPort,
		user:     cfg.SMTPUser,
		pass:     cfg.SMTPPass,
		fromAddr: addr,
		fromName: name,
		appURL:   cfg.AppURL,
	}
}

func (s *EmailService) send(to, subject, body string) error {
	if s.host == "" || s.user == "" {
		return nil // skip if SMTP not configured
	}

	from := s.fromAddr
	if s.fromName != "" {
		from = fmt.Sprintf("%s <%s>", s.fromName, s.fromAddr)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	return smtp.SendMail(addr, auth, s.fromAddr, []string{to}, []byte(msg))
}

func (s *EmailService) SendConfirmation(to, reference, guestName, checkIn, checkOut, roomName string) error {
	subject := "Booking Confirmed — Hotel Lobby"
	body := fmt.Sprintf(`Hi %s,

Your booking has been confirmed!

Reference: %s
Room: %s
Check-in: %s
Check-out: %s

You can view or manage your reservation at:
%s/reservations/%s

Thank you for choosing Hotel Lobby!`, guestName, reference, roomName, checkIn, checkOut, s.appURL, reference)

	return s.send(to, subject, body)
}

func (s *EmailService) SendPaymentReceipt(to, reference, amount, method, status string) error {
	subject := "Payment Receipt — Hotel Lobby"
	body := fmt.Sprintf(`Hi,

Payment for reservation %s has been processed.

Amount: %s
Method: %s
Status: %s

Thank you for your payment!`, reference, amount, method, status)

	return s.send(to, subject, body)
}

func (s *EmailService) SendCancellationConfirmation(to, reference string) error {
	subject := "Booking Cancelled — Hotel Lobby"
	body := fmt.Sprintf(`Hi,

Your reservation %s has been cancelled. If a refund is due, it will be processed within 5-7 business days.

We hope to welcome you again soon!`, reference)

	return s.send(to, subject, body)
}

func (s *EmailService) SendOTP(to, otp string) error {
	subject := "Your Verification Code — Hotel Lobby"
	body := fmt.Sprintf(`Hi,

Your verification code is: %s

This code expires in 15 minutes. If you did not request this, please ignore this email.`, otp)

	return s.send(to, subject, body)
}

func (s *EmailService) SendAbandonedBooking(to, reference string) error {
	subject := "You left something behind — Complete Your Booking"
	body := fmt.Sprintf(`Hi,

You started a booking but didn't finish. Your room is on hold — complete your booking now before it expires.

Resume your booking:
%s/reservations/%s/resume`, s.appURL, reference)

	return s.send(to, subject, body)
}
