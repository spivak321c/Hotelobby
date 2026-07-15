package router

import "github.com/gofiber/fiber/v2"

func (r *Router) registerPaymentRoutes(api fiber.Router) {
	api.Post("/payments", r.Payment.ProcessPayment)
	api.Post("/payments/webhook", r.Payment.HandleWebhook)
	api.Post("/payments/webhook/crossmint", r.Payment.HandleCrossmintWebhook)
	api.Get("/payments/:reference", r.Payment.CheckPayment)
}
