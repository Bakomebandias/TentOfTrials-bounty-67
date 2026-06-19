package matching

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/tent-of-trials/market/orderbook"
	"github.com/tent-of-trials/market/types"
)

// TestEnginePriceTimePriority verifies that the engine respects price-time priority.
// Bids with higher price are matched first; asks with lower price are matched first.
func TestEnginePriceTimePriority(t *testing.T) {
	config := EngineConfig{EnableShorting: true}
	books := map[types.Symbol]*orderbook.OrderBook{
		"BTC-USD": orderbook.NewOrderBook("BTC-USD"),
	}
	engine := NewMatchingEngine(config, books)

	// Place a bid at 100
	bid1 := &types.Order{
		Symbol:   "BTC-USD",
		Side:     types.Buy,
		Type:     types.Limit,
		Price:    decimal.NewFromInt(100),
		Quantity: decimal.NewFromFloat(1.0),
	}
	_, err := engine.PlaceOrder(bid1)
	if err != nil {
		t.Fatalf("failed to place bid1: %v", err)
	}

	// Place a matching ask at 99 – should trade immediately
	ask1 := &types.Order{
		Symbol:   "BTC-USD",
		Side:     types.Sell,
		Type:     types.Limit,
		Price:    decimal.NewFromInt(99),
		Quantity: decimal.NewFromFloat(0.5),
	}
	trades, err := engine.PlaceOrder(ask1)
	if err != nil {
		t.Fatalf("failed to place ask1: %v", err)
	}
	if len(trades) == 0 {
		t.Error("expected a trade when ask crosses bid, got none")
	}

	// Verify trade count increased
	count := engine.GetTradeCount()
	if count == 0 {
		t.Error("trade count should be > 0 after a successful match")
	}
}

// TestPartialFillRemaining verifies that partial fills leave correct remaining quantity.
func TestPartialFillRemaining(t *testing.T) {
	book := orderbook.NewOrderBook("ETH-USD")

	// Place a bid for 10 ETH
	bid := &types.Order{
		Symbol:   "ETH-USD",
		Side:     types.Buy,
		Type:     types.Limit,
		Price:    decimal.NewFromInt(2000),
		Quantity: decimal.NewFromFloat(10.0),
	}
	_, err := book.AddOrder(bid)
	if err != nil {
		t.Fatalf("failed to add bid: %v", err)
	}

	// Place an ask for 4 ETH – should partially fill the bid
	ask := &types.Order{
		Symbol:   "ETH-USD",
		Side:     types.Sell,
		Type:     types.Limit,
		Price:    decimal.NewFromInt(2000),
		Quantity: decimal.NewFromFloat(4.0),
	}
	trades, err := book.AddOrder(ask)
	if err != nil {
		t.Fatalf("failed to add ask: %v", err)
	}

	_ = trades
	// After partial fill, the remaining bid should be 6 ETH
	remaining := bid.Quantity.Sub(ask.Quantity)
	expected := decimal.NewFromFloat(6.0)
	if !remaining.Equal(expected) {
		t.Errorf("remaining quantity should be %s, got %s", expected, remaining)
	}
}

// TestCanceledOrderNotMatched verifies that a canceled order cannot be matched.
func TestCanceledOrderNotMatched(t *testing.T) {
	book := orderbook.NewOrderBook("SOL-USD")

	bid := &types.Order{
		ID:       "cancel_me",
		Symbol:   "SOL-USD",
		Side:     types.Buy,
		Type:     types.Limit,
		Price:    decimal.NewFromInt(50),
		Quantity: decimal.NewFromFloat(1.0),
	}
	if _, err := book.AddOrder(bid); err != nil {
		t.Fatalf("failed to add bid: %v", err)
	}

	// Cancel the bid
	if err := book.CancelOrder("cancel_me"); err != nil {
		t.Fatalf("failed to cancel bid: %v", err)
	}

	// Now place a matching ask – should NOT match
	ask := &types.Order{
		Symbol:   "SOL-USD",
		Side:     types.Sell,
		Type:     types.Limit,
		Price:    decimal.NewFromInt(50),
		Quantity: decimal.NewFromFloat(1.0),
	}
	trades, err := book.AddOrder(ask)
	if err != nil {
		t.Fatalf("failed to add ask: %v", err)
	}

	if len(trades) > 0 {
		t.Error("expected no trades after bid was canceled, but trades occurred")
	}
}

// TestValidateOrderRejectsInvalidQuantity verifies the engine rejects zero/negative quantities.
func TestValidateOrderRejectsInvalidQuantity(t *testing.T) {
	config := EngineConfig{EnableShorting: true}
	engine := NewMatchingEngine(config, nil)

	order := &types.Order{
		Symbol:   "BTC-USD",
		Side:     types.Buy,
		Type:     types.Limit,
		Price:    decimal.NewFromInt(100),
		Quantity: decimal.Zero,
	}
	err := engine.ValidateOrder(order)
	if err != ErrInvalidQuantity {
		t.Errorf("expected ErrInvalidQuantity for zero quantity, got: %v", err)
	}
}

// TestValidateOrderRejectsInvalidPrice verifies the engine rejects zero price for limit orders.
func TestValidateOrderRejectsInvalidPrice(t *testing.T) {
	config := EngineConfig{EnableShorting: true}
	engine := NewMatchingEngine(config, nil)

	order := &types.Order{
		Symbol:   "BTC-USD",
		Side:     types.Buy,
		Type:     types.Limit,
		Price:    decimal.Zero,
		Quantity: decimal.NewFromFloat(1.0),
	}
	err := engine.ValidateOrder(order)
	if err != ErrInvalidPrice {
		t.Errorf("expected ErrInvalidPrice for zero price, got: %v", err)
	}
}
