package router

import "github.com/gofiber/fiber/v2"

func (r *Router) registerRoomRoutes(api fiber.Router) {
	api.Get("/room-types", r.Room.ListRoomTypes)
	api.Get("/room-types/:id", r.Room.GetRoomType)
	api.Get("/rooms", r.Room.ListRooms)
	api.Get("/rooms/:id", r.Room.GetRoom)
	api.Get("/rooms/:id/images", r.Room.ListImages)
	api.Get("/rooms/:id/availability", r.Room.CheckAvailability)
}
