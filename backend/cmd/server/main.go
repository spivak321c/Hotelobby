package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"path/filepath"
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

//go:embed static
var staticFS embed.FS

var staticMIME = map[string]string{
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "application/javascript; charset=utf-8",
	".json": "application/json",
	".svg":  "image/svg+xml",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".ico":  "image/x-icon",
	".webp": "image/webp",
	".woff": "font/woff",
	".woff2": "font/woff2",
}

func contentType(path string) string {
	if ct, ok := staticMIME[filepath.Ext(path)]; ok {
		return ct
	}
	if ct := mime.TypeByExtension(filepath.Ext(path)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

func serveFile(c *fiber.Ctx, subFS fs.FS, path string) error {
	f, err := subFS.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	if stat.IsDir() {
		return serveFile(c, subFS, path+"/index.html")
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	c.Set("Content-Type", contentType(path))
	c.Set("Cache-Control", "public, max-age=3600")
	return c.Send(data)
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

	subFS, _ := fs.Sub(staticFS, "static")

	app.Use("/*", func(c *fiber.Ctx) error {
		path := c.Path()[1:]
		if path == "" {
			path = "index.html"
		}

		err := serveFile(c, subFS, path)
		if err != nil {
			err = serveFile(c, subFS, "200.html")
			if err != nil {
				return c.Status(404).SendString("Not found")
			}
		}
		return nil
	})

	log.Printf("starting server on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
