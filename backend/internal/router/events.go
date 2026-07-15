package router

import "github.com/gofiber/fiber/v2"

func (r *Router) registerEventRoutes(api fiber.Router) {
	// SSE events (public + auth)
	api.Get("/events", r.Admin.SSEEvents)
}
