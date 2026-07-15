package router

import (
	"time"

	"hotel_lobby/internal/handlers"
	"hotel_lobby/internal/middleware"
	"hotel_lobby/internal/repositories"
	"hotel_lobby/internal/services"
	"hotel_lobby/internal/sse"

	"github.com/gofiber/fiber/v2"
)

type Router struct {
	App         *fiber.App
	Auth        *handlers.AuthHandler
	Room        *handlers.RoomHandler
	Reservation *handlers.ReservationHandler
	Customer    *handlers.CustomerHandler
	Payment     *handlers.PaymentHandler
	Admin       *handlers.AdminHandler
	AuthService *services.AuthService

	globalLimiter *middleware.RateLimiter
	authLimiter   *middleware.RateLimiter
}

func New(
	app *fiber.App,
	authService *services.AuthService,
	roomService *services.RoomService,
	reservationService *services.ReservationService,
	customerService *services.CustomerService,
	paymentService *services.PaymentService,
	bookingService *services.BookingService,
	inventoryService *services.InventoryService,
	imageService *services.ImageService,
	sseHub *sse.Hub,
	customerRepo repositories.CustomerRepository,
	adminRepo repositories.AdminRepository,
) *Router {
	globalLimiter := middleware.NewRateLimiter(100, time.Minute)
	authLimiter := middleware.NewRateLimiter(5, time.Minute)
	globalLimiter.Cleanup(5 * time.Minute)

	return &Router{
		App:         app,
		Auth:        handlers.NewAuthHandler(authService),
		Room:        handlers.NewRoomHandler(roomService),
		Reservation: handlers.NewReservationHandler(reservationService),
		Customer:    handlers.NewCustomerHandler(customerService),
		Payment:     handlers.NewPaymentHandler(paymentService),
		Admin:       handlers.NewAdminHandler(roomService, reservationService, bookingService, inventoryService, authService, imageService, customerRepo, adminRepo, sseHub),
		AuthService: authService,
		globalLimiter: globalLimiter,
		authLimiter:   authLimiter,
	}
}

// RegisterAll groups and registers all domain-specific routes.
func (r *Router) RegisterAll() {
	api := r.App.Group("/api")
	api.Use(r.globalLimiter.Middleware)

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Domain routes
	r.registerAuthRoutes(api)
	r.registerRoomRoutes(api)
	r.registerReservationRoutes(api)
	r.registerPaymentRoutes(api)
	r.registerCustomerRoutes(api)
	r.registerAdminRoutes(api)
	r.registerEventRoutes(api)
}
