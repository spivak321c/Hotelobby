package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"hotel_lobby/internal/database"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	b, err := os.ReadFile("./.env")
	if err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
				parts := strings.SplitN(line, "=", 2)
				os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	db, err := database.NewPool(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	now := time.Now()

	// ── Clean up existing data (respecting FK order) ─────────
	log.Println("cleaning existing data...")
	db.Exec(ctx, `DELETE FROM payment`)
	db.Exec(ctx, `DELETE FROM booking`)
	db.Exec(ctx, `DELETE FROM reservation`)
	db.Exec(ctx, `DELETE FROM room_image`)
	db.Exec(ctx, `DELETE FROM room`)
	db.Exec(ctx, `DELETE FROM room_pricing`)
	db.Exec(ctx, `DELETE FROM room_type_inventory`)
	db.Exec(ctx, `DELETE FROM room_type`)
	log.Println("cleanup done")

	// ── Passwords ──────────────────────────────────────────────
	adminPass, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	custPass, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// ── Super Admin ────────────────────────────────────────────
	var superAdminID string
	err = db.QueryRow(ctx,
		`INSERT INTO admin (email, password_hash, full_name, role, is_active)
		 VALUES ($1,$2,$3,$4,true) ON CONFLICT (email) DO UPDATE SET full_name=EXCLUDED.full_name RETURNING id`,
		"admin@thelobby.com", string(adminPass), "Super Admin", "super_admin",
	).Scan(&superAdminID)
	if err != nil {
		log.Fatal("super admin:", err)
	}
	log.Println("super admin ID:", superAdminID)

	// ── Manager Admin ──────────────────────────────────────────
	var managerID string
	db.QueryRow(ctx,
		`INSERT INTO admin (email, password_hash, full_name, role, is_active)
		 VALUES ($1,$2,$3,$4,true) ON CONFLICT (email) DO UPDATE SET full_name=EXCLUDED.full_name RETURNING id`,
		"manager@thelobby.com", string(adminPass), "Sarah Manager", "manager",
	).Scan(&managerID)

	// ── Front Desk Admin ───────────────────────────────────────
	var frontDeskID string
	db.QueryRow(ctx,
		`INSERT INTO admin (email, password_hash, full_name, role, is_active)
		 VALUES ($1,$2,$3,$4,true) ON CONFLICT (email) DO UPDATE SET full_name=EXCLUDED.full_name RETURNING id`,
		"frontdesk@thelobby.com", string(adminPass), "Alex FrontDesk", "front_desk",
	).Scan(&frontDeskID)

	// ── Customers ──────────────────────────────────────────────
	customers := []struct{ email, name, phone string }{
		{"john@example.com", "John Smith", "+1-555-0101"},
		{"jane@example.com", "Jane Doe", "+1-555-0102"},
		{"mike@example.com", "Mike Johnson", "+1-555-0103"},
		{"sarah@example.com", "Sarah Williams", "+1-555-0104"},
		{"david@example.com", "David Brown", "+1-555-0105"},
	}

	custIDs := make([]string, len(customers))
	for i, c := range customers {
		var id string
		err := db.QueryRow(ctx,
			`INSERT INTO customer (email, password_hash, full_name, phone)
			 VALUES ($1,$2,$3,$4) ON CONFLICT (email) DO UPDATE SET full_name=EXCLUDED.full_name RETURNING id`,
			c.email, string(custPass), c.name, c.phone,
		).Scan(&id)
		if err != nil {
			log.Printf("customer %s: %v", c.email, err)
			continue
		}
		custIDs[i] = id
	}
	log.Println("customers seeded:", len(customers))

	// ── Room Types ─────────────────────────────────────────────
	typeIDs := make([]string, 3)
	types := []struct{ name, desc string; daily, hourly float64; occupancy int; featured bool }{
		{"Standard Double", "Comfortable double room with city view, en-suite bathroom, and modern amenities", 200, 50, 2, true},
		{"Deluxe Suite", "Spacious suite with premium amenities, living area, and panoramic views", 400, 100, 3, true},
		{"Presidential Villa", "Ultimate luxury with private terrace, butler service, and exclusive facilities", 800, 200, 4, true},
	}
	for i, t := range types {
		var id string
		db.QueryRow(ctx,
			`INSERT INTO room_type (name, description, base_rate_daily, base_rate_hourly, max_occupancy, is_featured)
			 VALUES ($1,$2,$3,$4,$5,$6)
			 ON CONFLICT DO NOTHING
			 RETURNING id`,
			t.name, t.desc, t.daily, t.hourly, t.occupancy, t.featured,
		).Scan(&id)
		if id == "" {
			// Already exists, fetch it
			db.QueryRow(ctx, `SELECT id FROM room_type WHERE name=$1`, t.name).Scan(&id)
		}
		typeIDs[i] = id
	}
	log.Println("room_type IDs:", typeIDs)

	// ── Rooms (4 per type = 12 total) ─────────────────────────
	roomIDs := make([]string, 12)
	for i := 0; i < 12; i++ {
		typeIdx := i / 4
		roomNum := fmt.Sprintf("%d%02d", typeIdx+1, (i%4)+1) // 101-112
		var id string
		err := db.QueryRow(ctx,
			`INSERT INTO room (room_type_id, room_number, status)
			 VALUES ($1,$2,'active')
			 ON CONFLICT (room_number) DO UPDATE SET status='active'
			 RETURNING id`,
			typeIDs[typeIdx], roomNum,
		).Scan(&id)
		if err != nil {
			log.Printf("room %s: %v", roomNum, err)
			continue
		}
		roomIDs[i] = id
	}
	log.Println("rooms seeded:", len(roomIDs))

	// ── Room Images (primary image per room) ───────────────────
	images := []string{
		"https://images.unsplash.com/photo-1631049307264-da0ec9d70304?w=800",
		"https://images.unsplash.com/photo-1590490360182-c33d57733427?w=800",
		"https://images.unsplash.com/photo-1582719508461-905c673771fd?w=800",
		"https://images.unsplash.com/photo-1591088398332-8a7791972843?w=800",
		"https://images.unsplash.com/photo-1564078516393-cf04bd966897?w=800",
		"https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=800",
	}
	for i, rid := range roomIDs {
		if rid == "" {
			continue
		}
		imgURL := images[i%len(images)]
		db.Exec(ctx,
			`INSERT INTO room_image (room_id, url, sort_order, is_primary)
			 VALUES ($1,$2,0,true) ON CONFLICT DO NOTHING`,
			rid, imgURL,
		)
	}

	// ── Room Pricing Overrides ─────────────────────────────────
	for _, tid := range typeIDs {
		if tid == "" {
			continue
		}
		// Get the room type's base rates
		var baseDaily, baseHourly float64
		db.QueryRow(ctx, `SELECT base_rate_daily, COALESCE(base_rate_hourly, base_rate_daily/4) FROM room_type WHERE id=$1`, tid).Scan(&baseDaily, &baseHourly)

		// Create a pricing override for next 30 days
		today := now.Format("2006-01-02")
		future := now.AddDate(0, 1, 0).Format("2006-01-02")
		rangeStr := fmt.Sprintf("[%s,%s)", today, future)
		db.Exec(ctx,
			`INSERT INTO room_pricing (room_type_id, rate_type, rate, effective_range)
			 VALUES ($1,'daily',$2,$3::daterange)
			 ON CONFLICT DO NOTHING`,
			tid, baseDaily*1.1, rangeStr,
		)
		db.Exec(ctx,
			`INSERT INTO room_pricing (room_type_id, rate_type, rate, effective_range)
			 VALUES ($1,'hourly',$2,$3::daterange)
			 ON CONFLICT DO NOTHING`,
			tid, baseHourly*1.1, rangeStr,
		)
	}

	// ── Inventory (60 days, bulk insert) ────────────────────────
	startDate := time.Now().Truncate(24 * time.Hour)
	for _, tid := range typeIDs {
		if tid == "" {
			continue
		}
		// Build a list of (room_type_id, date) pairs using generate_series
		_, err := db.Exec(ctx, `
			INSERT INTO room_type_inventory (room_type_id, date, total_rooms, booked_rooms)
			SELECT $1::uuid, generate_series($2::date, $2::date + 59, '1 day')::date, 4, 0
			ON CONFLICT (room_type_id, date) DO UPDATE SET total_rooms = 4
		`, tid, startDate.Format("2006-01-02"))
		if err != nil {
			log.Printf("inventory for type %s: %v", tid[:8], err)
		} else {
			log.Printf("inventory for type %s: 60 days ok", tid[:8])
		}
	}

	// ── Reservations & Bookings ────────────────────────────────
	type bookingSeed struct {
		guestName, guestEmail, guestPhone string
		custIdx                           int
		status                            string
		bookingType                       string
		daysFromNow                       int
		durationDays                      int
		roomIdx                           int
	}

	reservations := []bookingSeed{
		// Confirmed bookings
		{"John Smith", "john@example.com", "+1-555-0101", 0, "confirmed", "daily", 1, 3, 0},
		{"Jane Doe", "jane@example.com", "+1-555-0102", 1, "confirmed", "daily", 2, 2, 4},
		{"Mike Johnson", "mike@example.com", "+1-555-0103", 2, "confirmed", "hourly", 0, 0, 8},
		// Pending bookings
		{"Sarah Williams", "sarah@example.com", "+1-555-0104", 3, "pending", "daily", 5, 1, 1},
		{"David Brown", "david@example.com", "+1-555-0105", 4, "pending", "daily", 7, 2, 5},
		// Cancelled booking
		{"John Smith", "john@example.com", "+1-555-0101", 0, "cancelled", "daily", 10, 2, 2},
		// Past completed booking
		{"Jane Doe", "jane@example.com", "+1-555-0102", 1, "confirmed", "daily", -7, 3, 4},
		// More confirmed
		{"Mike Johnson", "mike@example.com", "+1-555-0103", 2, "confirmed", "daily", 3, 1, 9},
		{"Sarah Williams", "sarah@example.com", "+1-555-0104", 3, "confirmed", "hourly", 1, 0, 3},
		{"David Brown", "david@example.com", "+1-555-0105", 4, "confirmed", "daily", 4, 2, 7},
	}

	for i, rs := range reservations {
		refCode := fmt.Sprintf("TLB-%04d-%s", i+1, strings.ToUpper(uuid.New().String()[:4]))

		var custID *string
		if rs.custIdx < len(custIDs) && custIDs[rs.custIdx] != "" {
			custID = &custIDs[rs.custIdx]
		}

		checkIn := now.AddDate(0, 0, rs.daysFromNow)
		checkOut := checkIn.AddDate(0, 0, rs.durationDays)
		if rs.bookingType == "hourly" {
			checkOut = checkIn.Add(4 * time.Hour)
		}

		var roomTypeID string
		if rs.roomIdx < len(roomIDs) && roomIDs[rs.roomIdx] != "" {
			db.QueryRow(ctx, `SELECT room_type_id FROM room WHERE id=$1`, roomIDs[rs.roomIdx]).Scan(&roomTypeID)
		}

		// Calculate amount
		var amount float64
		db.QueryRow(ctx, `SELECT base_rate_daily FROM room_type WHERE id=$1`, roomTypeID).Scan(&amount)
		if rs.bookingType == "hourly" {
			db.QueryRow(ctx, `SELECT COALESCE(base_rate_hourly, base_rate_daily/4) FROM room_type WHERE id=$1`, roomTypeID).Scan(&amount)
			amount *= 4
		} else {
			amount *= float64(rs.durationDays)
		}

		var resID string
		err := db.QueryRow(ctx,
			`INSERT INTO reservation (reference_code, customer_id, guest_email, guest_name, guest_phone, status, total_amount, currency)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,'USD') RETURNING id`,
			refCode, custID, rs.guestEmail, rs.guestName, rs.guestPhone, rs.status, amount,
		).Scan(&resID)
		if err != nil {
			log.Printf("reservation %d: %v", i+1, err)
			continue
		}

		// Create booking
		if rs.roomIdx < len(roomIDs) && roomIDs[rs.roomIdx] != "" {
			bookingStatus := rs.status
			if rs.status == "cancelled" {
				bookingStatus = "cancelled"
			}

			_, err := db.Exec(ctx,
				`INSERT INTO booking (reservation_id, room_id, room_type_id, booking_type, starts_at, ends_at, status, amount)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
				resID, roomIDs[rs.roomIdx], roomTypeID, rs.bookingType,
				checkIn, checkOut, bookingStatus, amount,
			)
			if err != nil {
				log.Printf("booking for reservation %s: %v", refCode, err)
			}

			// Create payment for confirmed/completed reservations
			if rs.status == "confirmed" {
				providerRef := fmt.Sprintf("pay_%s", uuid.New().String()[:12])
				_, err := db.Exec(ctx,
					`INSERT INTO payment (reservation_id, provider, provider_reference, status, amount, currency)
					 VALUES ($1,'paystack',$2,'succeeded',$3,'USD')`,
					resID, providerRef, amount,
				)
				if err != nil {
					log.Printf("payment for reservation %s: %v", refCode, err)
				}
			}

			log.Printf("  reservation %s: %s (%s) %s %s → %s = $%.2f [%s]",
				refCode, rs.guestName, rs.bookingType,
				checkIn.Format("Jan 2"), checkOut.Format("Jan 2"),
				roomIDs[rs.roomIdx][:8], amount, rs.status)
		}
	}

	// ── Update inventory booked counts for confirmed bookings ──
	_, err = db.Exec(ctx, `
		UPDATE room_type_inventory rti
		SET booked_rooms = (
			SELECT COUNT(*)::int
			FROM booking b
			WHERE b.room_type_id = rti.room_type_id
			  AND b.status IN ('pending','confirmed','checked_in')
			  AND b.starts_at::date <= rti.date
			  AND b.ends_at::date > rti.date
		)
		WHERE rti.date >= $1::date
	`, startDate.Format("2006-01-02"))
	if err != nil {
		log.Printf("inventory update: %v", err)
	} else {
		log.Println("inventory booked counts updated")
	}

	log.Println("═══════════════════════════════════════")
	log.Println("  Seed completed successfully!")
	log.Println("═══════════════════════════════════════")
	log.Println("")
	log.Println("  Admin Login:")
	log.Println("    Email:    admin@thelobby.com")
	log.Println("    Password: admin123")
	log.Println("")
	log.Println("  Customer Login:")
	log.Println("    Email:    john@example.com")
	log.Println("    Password: password123")
	log.Println("")
	log.Println("  Rooms: 12 (4 Standard, 4 Deluxe, 4 Presidential)")
	log.Println("  Reservations: 10 (6 confirmed, 2 pending, 1 cancelled, 1 completed)")
	log.Println("═══════════════════════════════════════")
}
