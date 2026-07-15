package router

import (
	"hotel_lobby/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func (r *Router) registerCustomerRoutes(api fiber.Router) {
	customer := api.Group("/customer")
	customer.Use(middleware.AuthMiddleware(r.AuthService))

	customer.Get("/profile", r.Customer.GetProfile)
	customer.Put("/profile", r.Customer.UpdateProfile)
	customer.Get("/reservations", r.Customer.ListReservations)
	customer.Get("/reservations/:id", r.Customer.GetReservation)
}
