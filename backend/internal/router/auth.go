package router

import "github.com/gofiber/fiber/v2"

func (r *Router) registerAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")
	auth.Use(r.authLimiter.Middleware)
	auth.Post("/register", r.Auth.Register)
	auth.Post("/login", r.Auth.Login)
	auth.Post("/admin/login", r.Auth.AdminLogin)
	auth.Post("/refresh", r.Auth.Refresh)
	auth.Post("/logout", r.Auth.Logout)
}
