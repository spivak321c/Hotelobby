package main

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
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

var staticDir string

var mimeTypes = map[string]string{
	".html":  "text/html; charset=utf-8",
	".css":   "text/css; charset=utf-8",
	".js":    "application/javascript; charset=utf-8",
	".json":  "application/json",
	".svg":   "image/svg+xml",
	".png":   "image/png",
	".jpg":   "image/jpeg",
	".jpeg":  "image/jpeg",
	".gif":   "image/gif",
	".ico":   "image/x-icon",
	".webp":  "image/webp",
	".woff":  "font/woff",
	".woff2": "font/woff2",
	".ttf":   "font/ttf",
	".txt":   "text/plain; charset=utf-8",
}

func getContentType(path string) string {
	if ct, ok := mimeTypes[filepath.Ext(path)]; ok {
		return ct
	}
	if ct := mime.TypeByExtension(filepath.Ext(path)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

func staticHandler(c *fiber.Ctx) error {
	path := c.Path()[1:]
	if path == "" {
		path = "200.html"
	}

	// Resolve and sanitize: prevent directory traversal
	clean := filepath.Clean("/" + path)
	clean = strings.TrimPrefix(clean, "/")

	fullPath := filepath.Join(staticDir, clean)
	if !strings.HasPrefix(fullPath, staticDir) {
		return c.Status(403).SendString("Forbidden")
	}

	// Try exact file
	data, err := os.ReadFile(fullPath)
	if err == nil {
		c.Set("Content-Type", getContentType(clean))
		return c.Send(data)
	}

	// Try path.html (pre-rendered routes: /rooms -> rooms.html)
	data, err = os.ReadFile(fullPath + ".html")
	if err == nil {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(data)
	}

	// Try path/index.html
	data, err = os.ReadFile(filepath.Join(fullPath, "index.html"))
	if err == nil {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(data)
	}

	// SPA fallback
	data, err = os.ReadFile(filepath.Join(staticDir, "200.html"))
	if err == nil {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(data)
	}

	return c.Status(404).SendString("Not found")
}

func main() {
	godotenv.Load()
	ctx := context.Background()
	cfg := config.Load()

	// Static directory: env var STATIC_DIR or default to ./static relative to CWD
	staticDir = os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "static"
	}
	staticDir, _ = filepath.Abs(staticDir)

	// Log what we found for debugging
	if entries, err := os.ReadDir(staticDir); err != nil {
		log.Printf("WARNING: static dir %s not found: %v", staticDir, err)
	} else {
		log.Printf("static dir %s: %d entries", staticDir, len(entries))
	}

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

	appRouter := router.New(app, authService, roomService, reservationService, customerService, paymentService, bookingService, inventoryService, imageService, sseHub, customerRepo, adminRepo)
	appRouter.RegisterAll()

	app.Get("/", staticHandler)
	app.Use("/*", staticHandler)

	log.Printf("starting server on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
