package sse

import (
	"fmt"
	"sync"
	"testing"
)

func TestHub_SubscribeUnsubscribeUser(t *testing.T) {
	hub := NewHub()
	ch := make(chan Event, 1)

	hub.SubscribeUser("user1", ch)
	if len(hub.userCh["user1"]) != 1 {
		t.Fatalf("expected 1 channel for user1, got %d", len(hub.userCh["user1"]))
	}

	hub.UnsubscribeUser("user1", ch)
	if len(hub.userCh["user1"]) != 0 {
		t.Fatalf("expected 0 channels for user1, got %d", len(hub.userCh["user1"]))
	}
}

func TestHub_SubscribeUnsubscribeGuest(t *testing.T) {
	hub := NewHub()
	ch := make(chan Event, 1)

	hub.SubscribeGuest("guest1", ch)
	if len(hub.guestCh["guest1"]) != 1 {
		t.Fatalf("expected 1 channel for guest1, got %d", len(hub.guestCh["guest1"]))
	}

	hub.UnsubscribeGuest("guest1", ch)
	if len(hub.guestCh["guest1"]) != 0 {
		t.Fatalf("expected 0 channels for guest1, got %d", len(hub.guestCh["guest1"]))
	}
}

func TestHub_PublishToUser(t *testing.T) {
	hub := NewHub()
	ch := make(chan Event, 1)
	hub.SubscribeUser("user1", ch)

	event := Event{Type: "test", Payload: "data"}
	hub.PublishToUser("user1", event)

	select {
	case e := <-ch:
		if e.Type != "test" {
			t.Fatalf("expected test type, got %s", e.Type)
		}
	default:
		t.Fatal("expected event in channel")
	}
}

func TestHub_RaceConditions(t *testing.T) {
	hub := NewHub()
	var wg sync.WaitGroup

	numGoroutines := 100

	// Concurrent User Subscribers
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			userID := fmt.Sprintf("user%d", id%10)
			ch := make(chan Event, 10)
			hub.SubscribeUser(userID, ch)
			hub.PublishBookingUpdated(userID, "REF123", "confirmed")
			hub.UnsubscribeUser(userID, ch)
		}(i)
	}

	// Concurrent Guest Subscribers
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			sessionID := fmt.Sprintf("guest%d", id%10)
			ch := make(chan Event, 10)
			hub.SubscribeGuest(sessionID, ch)
			hub.PublishToGuest(sessionID, Event{Type: "ping"})
			hub.UnsubscribeGuest(sessionID, ch)
		}(i)
	}

	// Concurrent Broadcasters
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			hub.PublishAvailability("roomType1", "2026-07-08", []RoomAvailability{})
		}()
	}

	wg.Wait()
}
