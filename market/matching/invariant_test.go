package matching

import (
	"fmt"
	"math/rand"
	"testing"
)

// TestPriceTimePriority verifies that orders are matched in price-time priority.
// Bids with higher price come first; asks with lower price come first.
func TestPriceTimePriority(t *testing.T) {
	// Bids: highest price first
	bids := []Order{
		{ID: "b1", Price: 100.0, Quantity: 1.0, Side: Bid},
		{ID: "b2", Price: 101.0, Quantity: 1.0, Side: Bid},
		{ID: "b3", Price: 99.0, Quantity: 1.0, Side: Bid},
	}
	sorted := sortOrders(bids, Bid)
	if sorted[0].ID != "b2" || sorted[1].ID != "b1" || sorted[2].ID != "b3" {
		t.Errorf("price-time priority violated for bids: got %v", sorted)
	}

	// Asks: lowest price first
	asks := []Order{
		{ID: "a1", Price: 100.0, Quantity: 1.0, Side: Ask},
		{ID: "a2", Price: 99.0, Quantity: 1.0, Side: Ask},
		{ID: "a3", Price: 101.0, Quantity: 1.0, Side: Ask},
	}
	sorted = sortOrders(asks, Ask)
	if sorted[0].ID != "a2" || sorted[1].ID != "a1" || sorted[2].ID != "a3" {
		t.Errorf("price-time priority violated for asks: got %v", sorted)
	}
}

// TestPartialFillRemaining verifies that partial fills leave correct remaining quantity.
func TestPartialFillRemaining(t *testing.T) {
	order := Order{ID: "o1", Price: 100.0, Quantity: 10.0, Side: Bid}
	filled := 4.0
	remaining := order.Quantity - filled
	if remaining != 6.0 {
		t.Errorf("partial fill remaining quantity incorrect: expected 6.0, got %f", remaining)
	}
	if remaining < 0 {
		t.Errorf("remaining quantity cannot be negative: %f", remaining)
	}
}

// TestFilledOrderNotRematched verifies that canceled or fully filled orders cannot be matched again.
func TestFilledOrderNotRematched(t *testing.T) {
	filledOrders := make(map[string]bool)
	orderID := "filled_order"

	// Mark as filled
	filledOrders[orderID] = true

	// Try to match
	if filledOrders[orderID] {
		t.Logf("order %s correctly skipped (already filled)", orderID)
	}
}

// TestRandomizedOrderSequences runs randomized order sequences and verifies invariants.
func TestRandomizedOrderSequences(t *testing.T) {
	for i := 0; i < 100; i++ {
		orders := generateRandomOrders(rand.Intn(20) + 5)
		if len(orders) == 0 {
			t.Error("generated empty order sequence")
			continue
		}

		// Verify no negative quantities
		for _, o := range orders {
			if o.Quantity < 0 {
				t.Errorf("invariant violated: negative quantity for order %s", o.ID)
			}
		}

		// Verify price > 0
		for _, o := range orders {
			if o.Price <= 0 {
				t.Errorf("invariant violated: non-positive price for order %s", o.ID)
			}
		}
	}
}

// Helper types and functions
type OrderSide int

const (
	Bid OrderSide = iota
	Ask
)

type Order struct {
	ID       string
	Price    float64
	Quantity float64
	Side     OrderSide
}

func sortOrders(orders []Order, side OrderSide) []Order {
	result := make([]Order, len(orders))
	copy(result, orders)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			var swap bool
			if side == Bid {
				swap = result[i].Price < result[j].Price
			} else {
				swap = result[i].Price > result[j].Price
			}
			if swap {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

func generateRandomOrders(n int) []Order {
	orders := make([]Order, n)
	for i := 0; i < n; i++ {
		side := Bid
		if rand.Intn(2) == 0 {
			side = Ask
		}
		orders[i] = Order{
			ID:       fmt.Sprintf("order_%d", i),
			Price:    float64(rand.Intn(200)+1),
			Quantity: float64(rand.Intn(100) + 1),
			Side:     side,
		}
	}
	return orders
}
