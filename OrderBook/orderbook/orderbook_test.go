package orderbook

import (
	"testing"
)

func TestNewOrderBook(t *testing.T) {
	orderBook := testOrderBook
	// // try to restore before next operation
	// orderBook.Restore()

	if orderBook.VolumeAtPrice(Bid, Zero()).Cmp(Zero()) != 0 {
		t.Errorf("orderBook.VolumeAtPrice incorrect, got: %d, want: %d.", orderBook.VolumeAtPrice(Bid, Zero()), Zero())
	}

	if orderBook.BestAsk().Cmp(Zero()) != 0 {
		t.Errorf("orderBook.BestAsk incorrect, got: %d, want: %d.", orderBook.BestAsk(), Zero())
	}

	if orderBook.WorstBid().Cmp(Zero()) != 0 {
		t.Errorf("orderBook.WorstBid incorrect, got: %d, want: %d.", orderBook.WorstBid(), Zero())
	}

	if orderBook.WorstAsk().Cmp(Zero()) != 0 {
		t.Errorf("orderBook.WorstAsk incorrect, got: %d, want: %d.", orderBook.WorstAsk(), Zero())
	}

	if orderBook.BestBid().Cmp(Zero()) != 0 {
		t.Errorf("orderBook.BestBid incorrect, got: %d, want: %d.", orderBook.BestBid(), Zero())
	}
}

func TestOrderBook(t *testing.T) {
	orderBook := testOrderBook
	orderBook.Restore()

	limitOrders := make([]map[string]string, 0)

	dummyOrder := make(map[string]string)
	dummyOrder["type"] = "limit"
	dummyOrder["side"] = Ask
	dummyOrder["quantity"] = "5"
	dummyOrder["price"] = "101"
	dummyOrder["trade_id"] = "100"

	limitOrders = append(limitOrders, dummyOrder)

	dummyOrder1 := make(map[string]string)
	dummyOrder1["type"] = "limit"
	dummyOrder1["side"] = Ask
	dummyOrder1["quantity"] = "5"
	dummyOrder1["price"] = "103"
	dummyOrder1["trade_id"] = "101"

	limitOrders = append(limitOrders, dummyOrder1)

	dummyOrder2 := make(map[string]string)
	dummyOrder2["type"] = "limit"
	dummyOrder2["side"] = Ask
	dummyOrder2["quantity"] = "5"
	dummyOrder2["price"] = "101"
	dummyOrder2["trade_id"] = "102"

	limitOrders = append(limitOrders, dummyOrder2)

	dummyOrder7 := make(map[string]string)
	dummyOrder7["type"] = "limit"
	dummyOrder7["side"] = Ask
	dummyOrder7["quantity"] = "5"
	dummyOrder7["price"] = "101"
	dummyOrder7["trade_id"] = "103"

	limitOrders = append(limitOrders, dummyOrder7)

	dummyOrder3 := make(map[string]string)
	dummyOrder3["type"] = "limit"
	dummyOrder3["side"] = Bid
	dummyOrder3["quantity"] = "5"
	dummyOrder3["price"] = "99"
	dummyOrder3["trade_id"] = "100"

	limitOrders = append(limitOrders, dummyOrder3)

	dummyOrder4 := make(map[string]string)
	dummyOrder4["type"] = "limit"
	dummyOrder4["side"] = Bid
	dummyOrder4["quantity"] = "5"
	dummyOrder4["price"] = "98"
	dummyOrder4["trade_id"] = "101"
	limitOrders = append(limitOrders, dummyOrder4)

	dummyOrder5 := make(map[string]string)
	dummyOrder5["type"] = "limit"
	dummyOrder5["side"] = Bid
	dummyOrder5["quantity"] = "5"
	dummyOrder5["price"] = "99"
	dummyOrder5["trade_id"] = "102"

	limitOrders = append(limitOrders, dummyOrder5)

	dummyOrder6 := make(map[string]string)
	dummyOrder6["type"] = "limit"
	dummyOrder6["side"] = Bid
	dummyOrder6["quantity"] = "5"
	dummyOrder6["price"] = "97"
	dummyOrder6["trade_id"] = "103"

	limitOrders = append(limitOrders, dummyOrder6)

	// t.Logf("Limit Orders :%s", ToJSON(limitOrders))
	var trades []map[string]string
	var orderInBook map[string]string
	for _, order := range limitOrders {
		trades, orderInBook = orderBook.ProcessOrder(order, true)
	}
	// t.Logf("\nOrderBook :%s", orderBook.String(0))
	// return

	value := ToBigInt("101")
	if orderBook.BestAsk().Cmp(value) != 0 {
		t.Errorf("orderBook.BestAsk incorrect, got: %v, want: %v.", orderBook.BestAsk(), value)
	}

	value = ToBigInt("103")
	if orderBook.WorstAsk().Cmp(value) != 0 {
		t.Errorf("orderBook.WorstAsk incorrect, got: %v, want: %v.", orderBook.WorstAsk(), value)
	}

	value = ToBigInt("99")
	if orderBook.BestBid().Cmp(value) != 0 {
		t.Errorf("orderBook.BestBid incorrect, got: %v, want: %v.", orderBook.BestBid(), value)
	}

	value = ToBigInt("97")
	if orderBook.WorstBid().Cmp(value) != 0 {
		t.Errorf("orderBook.WorstBid incorrect, got: %v, want: %v.", orderBook.WorstBid(), value)
	}

	value = ToBigInt("15")
	pricePoint := ToBigInt("101")
	if orderBook.VolumeAtPrice(Ask, pricePoint).Cmp(value) != 0 {
		t.Errorf("orderBook.VolumeAtPrice incorrect, got: %v, want: %v.", orderBook.VolumeAtPrice(Ask, pricePoint), value)
	}

	//Submitting a limit order that crosses the opposing best price will result in a trade
	marketOrder := make(map[string]string)
	marketOrder["type"] = "limit"
	marketOrder["side"] = Bid
	marketOrder["quantity"] = "2"
	marketOrder["price"] = "102"
	marketOrder["trade_id"] = "109"

	trades, orderInBook = orderBook.ProcessOrder(marketOrder, true)

	if len(trades) > 0 {
		tradedPrice := trades[0]["price"]
		tradedQuantity := trades[0]["quantity"]

		if !(tradedPrice == "101" && tradedQuantity == "2" && len(orderInBook) == 0) {
			t.Errorf("orderBook.ProcessOrder incorrect")
		}
	}

	// t.Logf("\nOrderBook :%s", orderBook.String(0))
	// t.Logf("\nTrade :%s\nOrderInBook :%s", ToJSON(trades), ToJSON(orderInBook))
	// return

	// If a limit crosses but is only partially matched, the remaning volume will
	// be placed in the book as an outstanding order
	bigOrder := make(map[string]string)
	bigOrder["type"] = "limit"
	bigOrder["side"] = Bid
	bigOrder["quantity"] = "50"
	bigOrder["price"] = "102"
	bigOrder["trade_id"] = "110"

	trades, orderInBook = orderBook.ProcessOrder(bigOrder, true)

	if len(orderInBook) == 0 {
		t.Errorf("orderBook.ProcessOrder incorrect")
	}

	// t.Logf("\nOrderBook :%s", orderBook.String(0))
	// t.Logf("\nTrade :%s\nOrderInBook :%s", ToJSON(trades), ToJSON(orderInBook))
	// return

	// // Market orders only require that a user specifies a side (bid or ask), a quantity, and their unique trade id
	// marketOrder = make(map[string]string)
	// marketOrder["type"] = "market"
	// marketOrder["side"] = Ask
	// marketOrder["quantity"] = "20"
	// marketOrder["trade_id"] = "111"

	// trades, orderInBook = orderBook.ProcessOrder(marketOrder, true)

	// orderList := orderBook.Asks.MaxPriceList()
	// t.Logf("Best ask List : %s", orderList.String(0))
	// t.Log(orderBook.Asks.PriceTree)

	// orderBook.SetDebug(true)
	// Save to the database before exit
	orderBook.Commit()

	t.Logf("\nTrade :%s\nOrderInBook :%s", ToJSON(trades), ToJSON(orderInBook))
	t.Logf("\nOrderBook :%s", orderBook.String(0))

	// orderBook.Restore()
	// t.Logf("\nOrderBook :%s", orderBook.String(0))
}

func TestOrderBookRestore(t *testing.T) {
	orderBook := testOrderBook
	orderBook.SetDebug(true)

	orderBook.Restore()
	t.Logf("\nOrderBook :%s", orderBook.String(0))

	key := GetKeyFromString("10")
	order := orderBook.GetOrder(key)

	t.Logf("\nOrder : %s", order)
}
