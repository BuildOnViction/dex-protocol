package orderbook

import (
	"testing"
)

func TestNewOrderBook(t *testing.T) {
	orderBook := NewOrderBook(datadir)
	// // try to restore before next operation
	// orderBook.Restore()

	if orderBook.VolumeAtPrice(BID, Zero()).Cmp(Zero()) != 0 {
		t.Errorf("orderBook.VolumeAtPrice incorrect, got: %d, want: %d.", orderBook.VolumeAtPrice(BID, Zero()), Zero())
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
	orderBook := NewOrderBook(datadir)

	limitOrders := make([]map[string]string, 0)

	dummyOrder := make(map[string]string)
	dummyOrder["type"] = "limit"
	dummyOrder["side"] = ASK
	dummyOrder["quantity"] = "5"
	dummyOrder["price"] = "101"
	dummyOrder["trade_id"] = "100"

	limitOrders = append(limitOrders, dummyOrder)

	dummyOrder1 := make(map[string]string)
	dummyOrder1["type"] = "limit"
	dummyOrder1["side"] = ASK
	dummyOrder1["quantity"] = "5"
	dummyOrder1["price"] = "103"
	dummyOrder1["trade_id"] = "101"

	limitOrders = append(limitOrders, dummyOrder1)

	dummyOrder2 := make(map[string]string)
	dummyOrder2["type"] = "limit"
	dummyOrder2["side"] = ASK
	dummyOrder2["quantity"] = "5"
	dummyOrder2["price"] = "101"
	dummyOrder2["trade_id"] = "102"

	limitOrders = append(limitOrders, dummyOrder2)

	dummyOrder7 := make(map[string]string)
	dummyOrder7["type"] = "limit"
	dummyOrder7["side"] = ASK
	dummyOrder7["quantity"] = "5"
	dummyOrder7["price"] = "101"
	dummyOrder7["trade_id"] = "103"

	limitOrders = append(limitOrders, dummyOrder7)

	dummyOrder3 := make(map[string]string)
	dummyOrder3["type"] = "limit"
	dummyOrder3["side"] = BID
	dummyOrder3["quantity"] = "5"
	dummyOrder3["price"] = "99"
	dummyOrder3["trade_id"] = "100"

	limitOrders = append(limitOrders, dummyOrder3)

	dummyOrder4 := make(map[string]string)
	dummyOrder4["type"] = "limit"
	dummyOrder4["side"] = BID
	dummyOrder4["quantity"] = "5"
	dummyOrder4["price"] = "98"
	dummyOrder4["trade_id"] = "101"
	limitOrders = append(limitOrders, dummyOrder4)

	dummyOrder5 := make(map[string]string)
	dummyOrder5["type"] = "limit"
	dummyOrder5["side"] = BID
	dummyOrder5["quantity"] = "5"
	dummyOrder5["price"] = "99"
	dummyOrder5["trade_id"] = "102"

	limitOrders = append(limitOrders, dummyOrder5)

	dummyOrder6 := make(map[string]string)
	dummyOrder6["type"] = "limit"
	dummyOrder6["side"] = BID
	dummyOrder6["quantity"] = "5"
	dummyOrder6["price"] = "97"
	dummyOrder6["trade_id"] = "103"

	limitOrders = append(limitOrders, dummyOrder6)

	// t.Logf("Limit Orders :%s", ToJSON(limitOrders))

	for _, order := range limitOrders {
		orderBook.ProcessOrder(order, true)
	}

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
	if orderBook.VolumeAtPrice(ASK, pricePoint).Cmp(value) != 0 {
		t.Errorf("orderBook.VolumeAtPrice incorrect, got: %v, want: %v.", orderBook.VolumeAtPrice(ASK, pricePoint), value)
	}

	// //Submitting a limit order that crosses the opposing best price will result in a trade
	// marketOrder := make(map[string]string)
	// marketOrder["type"] = "limit"
	// marketOrder["side"] = BID
	// marketOrder["quantity"] = "2"
	// marketOrder["price"] = "102"
	// marketOrder["trade_id"] = "109"

	// trades, orderInBook := orderBook.ProcessOrder(marketOrder, true)

	// tradedPrice := trades[0]["price"]
	// tradedQuantity := trades[0]["quantity"]

	// if !(tradedPrice == "101" && tradedQuantity == "2" && len(orderInBook) == 0) {
	// 	t.Errorf("orderBook.ProcessOrder incorrect")
	// }

	// // If a limit crosses but is only partially matched, the remaning volume will
	// // be placed in the book as an outstanding order
	// bigOrder := make(map[string]string)
	// bigOrder["type"] = "limit"
	// bigOrder["side"] = BID
	// bigOrder["quantity"] = "50"
	// bigOrder["price"] = "102"
	// bigOrder["trade_id"] = "110"

	// trades, orderInBook = orderBook.ProcessOrder(bigOrder, true)

	// t.Logf("\nTrade :%s\nOrderInBook :%s", ToJSON(trades), ToJSON(orderInBook))

	// if !(len(orderInBook) != 0) {
	// 	t.Errorf("orderBook.ProcessOrder incorrect")
	// }

	// // Market orders only require that a user specifies a side (bid or ask), a quantity, and their unique trade id
	// marketOrder = make(map[string]string)
	// marketOrder["type"] = "market"
	// marketOrder["side"] = ASK
	// marketOrder["quantity"] = "20"
	// marketOrder["trade_id"] = "111"

	// trades, orderInBook = orderBook.ProcessOrder(marketOrder, true)

	// orderList := orderBook.Asks.MaxPriceList()
	// t.Logf("Best ask List : %s", orderList.String(0))
	t.Log(orderBook.Asks.PriceTree)
	t.Logf("\nOrderBook :%s", orderBook.String(0))

}
