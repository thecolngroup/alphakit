package backtest

import (
	"context"
	"testing"
	"time"

	"github.com/colngroup/zero2algo/broker"
	"github.com/colngroup/zero2algo/dec"
	"github.com/colngroup/zero2algo/market"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDealerProcessOrder(t *testing.T) {
	tests := []struct {
		name      string
		give      broker.Order
		wantOrder broker.Order
		wantState broker.OrderState
	}{
		{
			name: "ok: market order filled",
			give: broker.Order{
				Type: broker.Market,
				Size: dec.New(1),
			},
			wantOrder: broker.Order{
				FilledPrice: dec.New(10),
				FilledSize:  dec.New(1),
			},
			wantState: broker.OrderClosed,
		},
		{
			name: "ok: limit order filled",
			give: broker.Order{
				Type:       broker.Limit,
				LimitPrice: dec.New(8),
				Size:       dec.New(1),
			},
			wantOrder: broker.Order{
				FilledPrice: dec.New(8),
				FilledSize:  dec.New(1),
			},
			wantState: broker.OrderClosed,
		},
		{
			name: "ok: limit order open but not filled",
			give: broker.Order{
				Type:       broker.Limit,
				LimitPrice: dec.New(100),
				Size:       dec.New(1),
			},
			wantOrder: broker.Order{},
			wantState: broker.OrderOpen,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dealer := NewDealer()
			dealer.price = market.Kline{O: dec.New(8), H: dec.New(15), L: dec.New(5), C: dec.New(10)}
			act := dealer.processOrder(tt.give)
			assert.Equal(t, tt.wantOrder.FilledSize, act.FilledSize)
			assert.Equal(t, tt.wantOrder.FilledPrice, act.FilledPrice)
			assert.Equal(t, tt.wantState, act.State())
		})
	}
}

func TestDealerUpdatePosition(t *testing.T) {
	tests := []struct {
		name         string
		giveOrder    broker.Order
		givePosition broker.Position
		wantPosition broker.Position
		wantState    broker.PositionState
	}{
		{
			name:         "open new position",
			giveOrder:    broker.Order{Side: broker.Buy, FilledPrice: dec.New(10), FilledSize: dec.New(1)},
			givePosition: broker.Position{},
			wantPosition: broker.Position{
				Side:  broker.Buy,
				Price: dec.New(10),
				Size:  dec.New(1),
			},
			wantState: broker.PositionOpen,
		},
		{
			name:         "close existing position",
			giveOrder:    broker.Order{Side: broker.Sell, FilledPrice: dec.New(10), FilledSize: dec.New(1)},
			givePosition: broker.Position{Side: broker.Buy, Price: dec.New(10), Size: dec.New(1), OpenedAt: time.Now()},
			wantPosition: broker.Position{
				Side:  broker.Buy,
				Price: dec.New(10),
				Size:  dec.New(1),
			},
			wantState: broker.PositionClosed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dealer := NewDealer()
			act := dealer.updatePosition(tt.givePosition, tt.giveOrder)
			assert.Equal(t, tt.wantPosition.Side, act.Side)
			assert.Equal(t, tt.wantPosition.Price, act.Price)
			assert.Equal(t, tt.wantPosition.Size, act.Size)
			assert.Equal(t, tt.wantState, act.State())
		})
	}
}

func TestDealerGetLatestOrNewPosition(t *testing.T) {
	tests := []struct {
		name string
		give map[broker.DealID]broker.Position
		want broker.PositionState
	}{
		{
			name: "no positions",
			give: map[broker.DealID]broker.Position{},
			want: broker.PositionPending,
		},
		{
			name: "latest position is closed",
			give: map[broker.DealID]broker.Position{
				"1": {OpenedAt: time.Now()},
				"2": {ClosedAt: time.Now()},
			},
			want: broker.PositionPending,
		},
		{
			name: "latest position is open",
			give: map[broker.DealID]broker.Position{"1": {ID: "1", OpenedAt: time.Now()}},
			want: broker.PositionOpen,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dealer := NewDealer()
			dealer.positions = tt.give
			act := dealer.getLatestOrNewPosition()
			assert.Equal(t, tt.want, act.State())
		})
	}
}

func TestDealerPlaceOrder(t *testing.T) {
	tests := []struct {
		name string
		give broker.Order
		want error
	}{
		{
			name: "err: invalid side",
			give: broker.Order{
				Side: 0,
				Type: broker.Market,
				Size: dec.New(1),
			},
			want: ErrInvalidOrderState,
		},
		{
			name: "err: invalid type",
			give: broker.Order{
				Side: broker.Buy,
				Type: 0,
				Size: dec.New(1),
			},
			want: ErrInvalidOrderState,
		},
		{
			name: "err: invalid size",
			give: broker.Order{
				Side: broker.Buy,
				Type: broker.Market,
				Size: dec.New(0),
			},
			want: ErrInvalidOrderState,
		},
		{
			name: "err: no pending state",
			give: broker.Order{
				OpenedAt: time.Now(),
			},
			want: ErrInvalidOrderState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dealer Dealer
			act, _, err := dealer.PlaceOrder(context.Background(), tt.give)
			assert.ErrorIs(t, err, tt.want)
			assert.Nil(t, act)
		})
	}
}

func TestDealerReceivePrice(t *testing.T) {

	dealer := NewDealer()

	k1 := broker.NewIDWithTime(dealer.clock.Now())
	dealer.orders[k1] = broker.Order{ID: k1, Type: broker.Limit, LimitPrice: dec.New(15), OpenedAt: dealer.clock.Now()}

	k2 := broker.NewIDWithTime(dealer.clock.Now())
	dealer.orders[k2] = broker.Order{ID: k2, Type: broker.Limit, LimitPrice: dec.New(15), OpenedAt: dealer.clock.Now()}

	k3 := broker.NewIDWithTime(dealer.clock.Now())
	dealer.orders[k3] = broker.Order{ID: k3, Type: broker.Limit, LimitPrice: dec.New(10), OpenedAt: dealer.clock.Now()}

	price := market.Kline{
		Start: dealer.clock.Epoch().Add(time.Hour * 1),
		O:     dec.New(8),
		H:     dec.New(15),
		L:     dec.New(5),
		C:     dec.New(10)}

	dealer.ReceivePrice(context.Background(), price)

	t.Run("all open orders are closed", func(t *testing.T) {
		for _, v := range dealer.orders {
			if v.State() != broker.OrderClosed {
				assert.Fail(t, "expect all orders to be closed")
			}
		}
	})

	t.Run("orders are processed in order they were created", func(t *testing.T) {
		assert.True(t, dealer.orders[k1].ClosedAt.Before(dealer.orders[k2].ClosedAt))
		assert.True(t, dealer.orders[k2].ClosedAt.Before(dealer.orders[k3].ClosedAt))
	})
}

func TestDealerOpenOrder(t *testing.T) {
	dealer := NewDealer()
	order := dealer.openOrder(broker.Order{})
	assert.EqualValues(t, broker.OrderOpen, order.State())
}

func TestDealerFillOrder(t *testing.T) {
	dealer := NewDealer()
	order := dealer.fillOrder(broker.Order{}, dec.New(100))
	assert.EqualValues(t, broker.OrderFilled, order.State())
}

func TestDealerCloseOrder(t *testing.T) {
	dealer := NewDealer()
	order := dealer.closeOrder(broker.Order{})
	assert.EqualValues(t, broker.OrderClosed, order.State())
}

func TestMatchOrder(t *testing.T) {
	tests := []struct {
		name      string
		giveOrder broker.Order
		giveQuote market.Kline
		want      decimal.Decimal
	}{
		{
			name: "ok: match limit",
			giveOrder: broker.Order{
				Type:       broker.Limit,
				LimitPrice: dec.New(12),
			},
			giveQuote: market.Kline{O: dec.New(8), H: dec.New(15), L: dec.New(5), C: dec.New(10)},
			want:      dec.New(12),
		},
		{
			name: "ok: match limit lower bound inclusive",
			giveOrder: broker.Order{
				Type:       broker.Limit,
				LimitPrice: dec.New(5),
			},
			giveQuote: market.Kline{O: dec.New(8), H: dec.New(15), L: dec.New(5), C: dec.New(10)},
			want:      dec.New(5),
		},
		{
			name: "ok: match limit upper bound inclusive",
			giveOrder: broker.Order{
				Type:       broker.Limit,
				LimitPrice: dec.New(15),
			},
			giveQuote: market.Kline{O: dec.New(8), H: dec.New(15), L: dec.New(5), C: dec.New(10)},
			want:      dec.New(15),
		},
		{
			name: "ok: no match limit below lower bound",
			giveOrder: broker.Order{
				Type:       broker.Limit,
				LimitPrice: dec.New(2),
			},
			giveQuote: market.Kline{O: dec.New(8), H: dec.New(15), L: dec.New(5), C: dec.New(10)},
			want:      decimal.Decimal{},
		},
		{
			name: "ok: no match limit above upper bound",
			giveOrder: broker.Order{
				Type:       broker.Limit,
				LimitPrice: dec.New(100),
			},
			giveQuote: market.Kline{O: dec.New(8), H: dec.New(15), L: dec.New(5), C: dec.New(10)},
			want:      decimal.Decimal{},
		},
		{
			name: "ok: always match market on close price",
			giveOrder: broker.Order{
				Type: broker.Market,
			},
			giveQuote: market.Kline{O: dec.New(8), H: dec.New(15), L: dec.New(5), C: dec.New(10)},
			want:      dec.New(10),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act := matchOrder(tt.giveOrder, tt.giveQuote)
			assert.Equal(t, tt.want, act)
		})
	}
}

func TestCloseTime(t *testing.T) {
	interval := time.Hour * 4
	start1 := time.Now()
	start2 := start1.Add(interval)

	t.Run("valid start times", func(t *testing.T) {
		exp := start2.Add(interval)
		act := closeTime(start1, start2)
		assert.EqualValues(t, exp, act)
	})

	t.Run("start 1 is zero", func(t *testing.T) {
		act := closeTime(time.Time{}, start2)
		assert.EqualValues(t, start2, act)
	})
}

func TestProfit(t *testing.T) {
	tests := []struct {
		name string
		give broker.Position
		want decimal.Decimal
	}{
		{
			name: "buy side profit",
			give: broker.Position{
				Side:             broker.Buy,
				Price:            dec.New(10),
				Size:             dec.New(2),
				LiquidationPrice: dec.New(20),
			},
			want: dec.New(20),
		},
		{
			name: "sell side profit",
			give: broker.Position{
				Side:             broker.Sell,
				Price:            dec.New(100),
				Size:             dec.New(2),
				LiquidationPrice: dec.New(50),
			},
			want: dec.New(100),
		},
		{
			name: "buy side loss",
			give: broker.Position{
				Side:             broker.Buy,
				Price:            dec.New(10),
				Size:             dec.New(2),
				LiquidationPrice: dec.New(5),
			},
			want: dec.New(-10),
		},
		{
			name: "sell side loss",
			give: broker.Position{
				Side:             broker.Sell,
				Price:            dec.New(10),
				Size:             dec.New(2),
				LiquidationPrice: dec.New(20),
			},
			want: dec.New(-20),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act := profit(tt.give)
			assert.Equal(t, tt.want, act)
		})
	}
}
