package sse

import (
	"sync"
)

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Hub struct {
	mu      sync.RWMutex
	userCh  map[string][]chan Event // user_id -> subscriber channels
	guestCh map[string][]chan Event // guest_session -> subscriber channels
}

func NewHub() *Hub {
	return &Hub{
		userCh:  make(map[string][]chan Event),
		guestCh: make(map[string][]chan Event),
	}
}

func (h *Hub) SubscribeUser(userID string, ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.userCh[userID] = append(h.userCh[userID], ch)
}

func (h *Hub) SubscribeGuest(sessionID string, ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.guestCh[sessionID] = append(h.guestCh[sessionID], ch)
}

func (h *Hub) UnsubscribeUser(userID string, ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	chs := h.userCh[userID]
	for i, c := range chs {
		if c == ch {
			h.userCh[userID] = append(chs[:i], chs[i+1:]...)
			break
		}
	}
	if len(h.userCh[userID]) == 0 {
		delete(h.userCh, userID)
	}
}

func (h *Hub) UnsubscribeGuest(sessionID string, ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	chs := h.guestCh[sessionID]
	for i, c := range chs {
		if c == ch {
			h.guestCh[sessionID] = append(chs[:i], chs[i+1:]...)
			break
		}
	}
	if len(h.guestCh[sessionID]) == 0 {
		delete(h.guestCh, sessionID)
	}
}

func (h *Hub) PublishToUser(userID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.userCh[userID] {
		select {
		case ch <- event:
		default:
		}
	}
}

func (h *Hub) PublishToGuest(sessionID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.guestCh[sessionID] {
		select {
		case ch <- event:
		default:
		}
	}
}

type RoomAvailability struct {
	RoomID     string `json:"room_id"`
	RoomNumber string `json:"room_number"`
	Available  bool   `json:"available"`
}

func (h *Hub) PublishAvailability(roomTypeID, date string, rooms []RoomAvailability) {
	event := Event{
		Type: "availability",
		Payload: map[string]interface{}{
			"room_type_id": roomTypeID,
			"date":         date,
			"rooms":        rooms,
		},
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, chs := range h.userCh {
		for _, ch := range chs {
			select {
			case ch <- event:
			default:
			}
		}
	}
	for _, chs := range h.guestCh {
		for _, ch := range chs {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

func (h *Hub) PublishBookingUpdated(userID, reference, status string) {
	event := Event{
		Type: "booking-updated",
		Payload: map[string]interface{}{
			"reference": reference,
			"status":    status,
		},
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.userCh[userID] {
		select {
		case ch <- event:
		default:
		}
	}
}
