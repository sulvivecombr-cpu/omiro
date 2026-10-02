package main

import (
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
)

var queue []string
var queueMu sync.Mutex

func handleJoinQueue(c *Client) {
	// Never queue a client that still has a partner.
	releasePartner(c)

	queueMu.Lock()
	defer queueMu.Unlock()

	if slices.Contains(queue, c.ID) {
		return
	}
	queue = append(queue, c.ID)
	log.Println("added to queue:", c.ID)
	findMatch()
}

func removeFromQueue(id string) {
	queueMu.Lock()
	defer queueMu.Unlock()
	if i := slices.Index(queue, id); i >= 0 {
		queue = slices.Delete(queue, i, i+1)
		log.Println("removed from queue:", id)
	}
}

// releasePartner breaks the pairing (if any) and notifies the partner.
func releasePartner(c *Client) {
	clientsMu.Lock()
	partner := c.Partner
	c.Partner = nil
	if partner != nil && partner.Partner == c {
		partner.Partner = nil
	}
	clientsMu.Unlock()

	if partner != nil {
		partner.trySend(websocket.TextMessage, []byte(`{"op":"partner_disconnected"}`))
	}
}

func handleLeaveQueue(c *Client) {
	removeFromQueue(c.ID)
	releasePartner(c)
}

// findMatch must be called with queueMu held.
func findMatch() {
	for len(queue) >= 2 {
		clientsMu.RLock()
		c1, c2 := clients[queue[0]], clients[queue[1]]
		clientsMu.RUnlock()

		// Drop stale (disconnected) entries but keep live ones queued.
		if c1 == nil {
			queue = queue[1:]
			continue
		}
		if c2 == nil {
			queue = append(queue[:1], queue[2:]...)
			continue
		}
		queue = queue[2:]

		clientsMu.Lock()
		c1.Partner = c2
		c2.Partner = c1
		clientsMu.Unlock()

		log.Println("matched:", c1.ID, "<->", c2.ID)
		// c1 is the caller, c2 waits for the offer.
		c1.trySend(websocket.TextMessage, fmt.Appendf(nil,
			`{"op":"match_found","partner":"%s","should_call":true}`, c2.ID))
		c2.trySend(websocket.TextMessage, fmt.Appendf(nil,
			`{"op":"match_found","partner":"%s","should_call":false}`, c1.ID))
		return
	}
}
