package router

import "github.com/gofiber/fiber/v2"

func (r *Router) registerReservationRoutes(api fiber.Router) {
	reservations := api.Group("/reservations")
	reservations.Post("/", r.Reservation.Create)
	reservations.Get("/:reference", r.Reservation.Lookup)
	reservations.Post("/:reference/cancel/otp", r.Reservation.RequestCancelOTP)
	reservations.Post("/:reference/cancel", r.Reservation.Cancel)
}
