package router

import (
	"hotel_lobby/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func (r *Router) registerAdminRoutes(api fiber.Router) {
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(r.AuthService))

	// Common role guard for most admin endpoints
	roleGuard := middleware.RoleMiddleware("super_admin", "manager")

	// Room Types
	admin.Get("/room-types", roleGuard, r.Admin.ListRoomTypes)
	admin.Post("/room-types", roleGuard, r.Admin.CreateRoomType)
	admin.Put("/room-types/:id", roleGuard, r.Admin.UpdateRoomType)
	admin.Delete("/room-types/:id", roleGuard, r.Admin.DeleteRoomType)

	// Rooms
	admin.Get("/rooms", roleGuard, r.Admin.ListRooms)
	admin.Post("/rooms", roleGuard, r.Admin.CreateRoom)
	admin.Put("/rooms/:id", roleGuard, r.Admin.UpdateRoom)
	admin.Delete("/rooms/:id", roleGuard, r.Admin.DeleteRoom)

	// Room Images
	admin.Post("/rooms/:id/images", roleGuard, r.Admin.UploadImage)
	admin.Delete("/rooms/:id/images/:image_id", roleGuard, r.Admin.DeleteImage)
	admin.Put("/rooms/:id/images/reorder", roleGuard, r.Admin.ReorderImages)

	// Pricing
	admin.Get("/room-pricing", roleGuard, r.Admin.ListRoomPricing)
	admin.Post("/room-pricing", roleGuard, r.Admin.CreateRoomPricing)
	admin.Put("/room-pricing/:id", roleGuard, r.Admin.UpdateRoomPricing)
	admin.Delete("/room-pricing/:id", roleGuard, r.Admin.DeleteRoomPricing)

	// Inventory
	admin.Get("/inventory", roleGuard, r.Admin.GetInventory)
	admin.Put("/inventory", roleGuard, r.Admin.UpdateInventory)

	// Reservations (includes front_desk)
	reservationGuard := middleware.RoleMiddleware("super_admin", "manager", "front_desk")
	admin.Get("/reservations", reservationGuard, r.Admin.ListReservations)
	admin.Get("/reservations/:id", reservationGuard, r.Admin.GetReservation)
	admin.Put("/reservations/:id/status", reservationGuard, r.Admin.UpdateReservationStatus)
	admin.Post("/walk-in", reservationGuard, r.Admin.CreateWalkIn)

	// User Management (super_admin only)
	superAdminGuard := middleware.RoleMiddleware("super_admin")
	admin.Get("/customers", superAdminGuard, r.Admin.ListCustomers)
	admin.Get("/customers/:id", superAdminGuard, r.Admin.GetCustomer)
	admin.Get("/admins", superAdminGuard, r.Admin.ListAdmins)
	admin.Post("/admins", superAdminGuard, r.Admin.CreateAdmin)
	admin.Put("/admins/:id", superAdminGuard, r.Admin.UpdateAdmin)
	admin.Delete("/admins/:id", superAdminGuard, r.Admin.DeleteAdmin)

	// Reports
	admin.Get("/reports/bookings", roleGuard, r.Admin.BookingReport)
	admin.Get("/reports/occupancy", roleGuard, r.Admin.OccupancyReport)
	admin.Get("/reports/revenue", roleGuard, r.Admin.RevenueReport)
}
