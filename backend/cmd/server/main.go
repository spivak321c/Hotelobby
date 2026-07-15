package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"time"

	"hotel_lobby/internal/config"
	"hotel_lobby/internal/database"
	"hotel_lobby/internal/providers/cloudinary"
	"hotel_lobby/internal/repositories"
	"hotel_lobby/internal/router"
	"hotel_lobby/internal/services"
	"hotel_lobby/internal/sse"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"
)

//go:embed static/*
var staticFS embed.FS

func newStaticHandler() fiber.Handler {
	subFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("static embed sub: %v", err)
	}

	return func(c *fiber.Ctx) error {
		path := strings.TrimPrefix(c.Path(), "/")
		if path == "" {
			path = "index.html"
		}

		// Try exact path, path.html, path/index.html, then SPA fallback
		if data, err := fs.ReadFile(subFS, path); err == nil {
			c.Type(filepath.Ext(path))
			return c.Send(data)
		}

		if data, err := fs.ReadFile(subFS, path+".html"); err == nil {
			c.Type(".html")
			return c.Send(data)
		}

		if data, err := fs.ReadFile(subFS, path+"/index.html"); err == nil {
			c.Type(".html")
			return c.Send(data)
		}

		// SPA fallback
		data, err := fs.ReadFile(subFS, "200.html")
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}
		c.Type(".html")
		return c.Send(data)
	}
}

func main() {
	godotenv.Load()
	ctx := context.Background()
	cfg := config.Load()

	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	rdb, err := database.NewRedis(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	// Repos
	roomTypeRepo := repositories.NewRoomTypeRepo(db)
	roomRepo := repositories.NewRoomRepo(db)
	imageRepo := repositories.NewRoomImageRepo(db)
	pricingRepo := repositories.NewRoomPricingRepo(db)
	inventoryRepo := repositories.NewRoomTypeInventoryRepo(db)
	customerRepo := repositories.NewCustomerRepo(db)
	adminRepo := repositories.NewAdminRepo(db)
	reservationRepo := repositories.NewReservationRepo(db)
	bookingRepo := repositories.NewBookingRepo(db)
	paymentRepo := repositories.NewPaymentRepo(db)

	// Services
	authService := services.NewAuthService(customerRepo, adminRepo, cfg.JWTSecret)
	roomService := services.NewRoomService(roomTypeRepo, roomRepo, pricingRepo, inventoryRepo, imageRepo)
	emailService := services.NewEmailService(services.EmailConfig{
		SMTPServer: cfg.SMTPServer,
		SMTPPort:   cfg.SMTPPort,
		SMTPUser:   cfg.SMTPUser,
		SMTPPass:   cfg.SMTPPass,
		EmailFrom:  cfg.EmailFrom,
		AppURL:     cfg.AppURL,
	})
	sseHub := sse.NewHub()
	otpStore := services.NewRedisOTPStore(rdb)
	inventoryService := services.NewInventoryService(rdb, inventoryRepo)
	reservationService := services.NewReservationService(reservationRepo, bookingRepo, paymentRepo, roomRepo, inventoryRepo, pricingRepo, roomTypeRepo, otpStore, emailService, sseHub, inventoryService)
	customerService := services.NewCustomerService(customerRepo, reservationRepo)
	bookingService := services.NewBookingService(bookingRepo, roomRepo, reservationRepo)
	paymentService := services.NewPaymentService(cfg.PaystackSecretKey, cfg.PaystackPublicKey, cfg.PaystackWebhookSec, cfg.CrossmintAPIKey, cfg.CrossmintProjectID, cfg.CrossmintWebhook, paymentRepo, reservationRepo, bookingRepo, emailService)

	// Image/Cloudinary
	cloudinaryURL := fmt.Sprintf("cloudinary://%s:%s@%s", cfg.CloudAPIKey, cfg.CloudSecret, cfg.CloudName)
	cldClient, err := cloudinary.NewClient(cloudinaryURL)
	if err != nil {
		log.Fatalf("cloudinary init: %v", err)
	}
	imageService := services.NewImageService(cldClient, imageRepo)

	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowCredentials: true,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Idempotency-Key",
	}))

	// Router Setup
	appRouter := router.New(app, authService, roomService, reservationService, customerService, paymentService, bookingService, inventoryService, imageService, sseHub, customerRepo, adminRepo)
	appRouter.RegisterAll()

	// Static file serving with SPA fallback
	staticHandler := newStaticHandler()
	app.Get("/", staticHandler)
	app.All("/*", staticHandler)

	// Log embedded file count for debugging
	if entries, err := fs.ReadDir(staticFS, "static"); err == nil {
		log.Printf("embedded static files: %d entries", len(entries))
	} else {
		log.Printf("warning: no embedded static files: %v", err)
	}

	port := cfg.Port
	if port == "" {
		port = "8000"
	}

	log.Printf("starting server on :%s", port)
	log.Fatal(app.Listen(":" + port))
}
