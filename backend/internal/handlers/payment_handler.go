package handlers

import (
	"hotel_lobby/internal/middleware"
	"hotel_lobby/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
}

func NewPaymentHandler(ps *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: ps}
}

func (h *PaymentHandler) ProcessPayment(c *fiber.Ctx) error {
	var req ProcessPaymentRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	reservationID, _ := uuid.Parse(req.ReservationID)

	input := services.ProcessPaymentInput{
		ReservationID: reservationID,
		Method:        req.Method,
	}

	result, err := h.paymentService.ProcessPayment(c.Context(), input)
	if err != nil {
		switch err {
		case services.ErrInvalidPaymentMethod:
			return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("invalid_payment_method", err.Error()))
		case services.ErrPaymentFailed:
			return middleware.Respond(c, fiber.StatusPaymentRequired, middleware.Fail("payment_failed", err.Error()))
		default:
			return middleware.Respond(c, fiber.StatusInternalServerError, middleware.Fail("payment_error", err.Error()))
		}
	}

	return middleware.Respond(c, fiber.StatusCreated, middleware.OK(fiber.Map{
		"status":             result.Status,
		"provider_reference": result.ProviderReference,
	}))
}

func (h *PaymentHandler) HandleWebhook(c *fiber.Ctx) error {
	sig := c.Get("x-paystack-signature")
	if sig == "" {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("missing_signature", "missing x-paystack-signature header"))
	}

	body := c.Body()
	if len(body) == 0 {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("empty_body", "request body is empty"))
	}

	if err := h.paymentService.HandlePaystackWebhook(c.Context(), body, sig); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("webhook_error", err.Error()))
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *PaymentHandler) HandleCrossmintWebhook(c *fiber.Ctx) error {
	sig := c.Get("x-cmn-signature")
	if sig == "" {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("missing_signature", "missing x-cmn-signature header"))
	}

	body := c.Body()
	if len(body) == 0 {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("empty_body", "request body is empty"))
	}

	if err := h.paymentService.HandleCrossmintWebhook(c.Context(), body, sig); err != nil {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("webhook_error", err.Error()))
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *PaymentHandler) CheckPayment(c *fiber.Ctx) error {
	txRef := c.Params("reference")
	if txRef == "" {
		return middleware.Respond(c, fiber.StatusBadRequest, middleware.Fail("missing_reference", "reference is required"))
	}

	payment, err := h.paymentService.CheckPayment(c.Context(), txRef)
	if err != nil {
		return middleware.Respond(c, fiber.StatusNotFound, middleware.Fail("payment_not_found", "payment not found"))
	}

	return middleware.Respond(c, fiber.StatusOK, middleware.OK(fiber.Map{
		"id":                 payment.ID.String(),
		"reservation_id":     payment.ReservationID.String(),
		"amount":             payment.Amount,
		"provider":           payment.Provider,
		"status":             payment.Status,
		"provider_reference": payment.ProviderReference,
	}))
}
