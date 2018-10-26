package orderbook

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	lane "gopkg.in/oleiade/lane.v1"
)

const (
	// ASK : ask constant
	ASK = "ask"
	// BID : bid constant
	BID = "bid"
)

// OrderBook : list of orders
type OrderBook struct {
	Deque       *lane.Deque `json:"deque"`
	Bids        *OrderTree  `json:"bids"`
	Asks        *OrderTree  `json:"asks"`
	Time        int         `json:"time"`
	NextOrderID int         `json:"nextOrderID"`
}

// NewOrderBook : return new order book
func NewOrderBook() *OrderBook {
	deque := lane.NewDeque()
	bids := NewOrderTree()
	asks := NewOrderTree()
	return &OrderBook{
		Deque:       deque,
		Bids:        bids,
		Asks:        asks,
		Time:        0,
		NextOrderID: 0,
	}
}

func (orderBook *OrderBook) String(startDepth int) string {
	tabs := strings.Repeat("\t", startDepth)
	return fmt.Sprintf("{\n\t%sBids: %s\n\t%sAsks: %s\n\t%sTime: %d\n\t%sNextOrderID: %d\n%s}\n",
		tabs, orderBook.Bids.String(startDepth+1), tabs, orderBook.Asks.String(startDepth+1), tabs, orderBook.Time, tabs, orderBook.NextOrderID, tabs)
}

// UpdateTime : update time for order book
func (orderBook *OrderBook) UpdateTime() {
	orderBook.Time++
}

// BestBid : get the best bid of the order book
func (orderBook *OrderBook) BestBid() (value decimal.Decimal) {
	return orderBook.Bids.MaxPrice()
}

// BestAsk : get the best ask of the order book
func (orderBook *OrderBook) BestAsk() (value decimal.Decimal) {
	return orderBook.Asks.MinPrice()
}

// WorstBid : get the worst bid of the order book
func (orderBook *OrderBook) WorstBid() (value decimal.Decimal) {
	return orderBook.Bids.MinPrice()
}

// WorstAsk : get the worst ask of the order book
func (orderBook *OrderBook) WorstAsk() (value decimal.Decimal) {
	return orderBook.Asks.MaxPrice()
}

// processMarketOrder : process the market order
func (orderBook *OrderBook) processMarketOrder(quote map[string]string, verbose bool) []map[string]string {
	var trades []map[string]string
	quantityToTrade, _ := decimal.NewFromString(quote["quantity"])
	side := quote["side"]
	var newTrades []map[string]string

	if side == BID {
		for quantityToTrade.GreaterThan(decimal.Zero) && orderBook.Asks.Length() > 0 {
			bestPriceAsks := orderBook.Asks.MinPriceList()
			quantityToTrade, newTrades = orderBook.processOrderList(ASK, bestPriceAsks, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
		}
	} else if side == ASK {
		for quantityToTrade.GreaterThan(decimal.Zero) && orderBook.Bids.Length() > 0 {
			bestPriceBids := orderBook.Bids.MaxPriceList()
			quantityToTrade, newTrades = orderBook.processOrderList(BID, bestPriceBids, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
		}
	}
	return trades
}

// processLimitOrder : process the limit order, can change the quote
// If not care for performance, we should make a copy of quote to prevent further reference problem
func (orderBook *OrderBook) processLimitOrder(quote map[string]string, verbose bool) ([]map[string]string, map[string]string) {
	var trades []map[string]string
	quantityToTrade, _ := decimal.NewFromString(quote["quantity"])
	side := quote["side"]
	price, _ := decimal.NewFromString(quote["price"])
	var newTrades []map[string]string

	var orderInBook map[string]string

	if side == BID {
		minPrice := orderBook.Asks.MinPrice()
		for quantityToTrade.GreaterThan(decimal.Zero) && orderBook.Asks.Length() > 0 && price.GreaterThanOrEqual(minPrice) {
			bestPriceAsks := orderBook.Asks.MinPriceList()
			quantityToTrade, newTrades = orderBook.processOrderList(ASK, bestPriceAsks, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
			minPrice = orderBook.Asks.MinPrice()
		}

		if quantityToTrade.GreaterThan(decimal.Zero) {
			quote["order_id"] = strconv.Itoa(orderBook.NextOrderID)
			quote["quantity"] = quantityToTrade.String()
			orderBook.Bids.InsertOrder(quote)
			orderInBook = quote
		}

	} else if side == "ask" {
		maxPrice := orderBook.Bids.MaxPrice()
		for quantityToTrade.GreaterThan(decimal.Zero) && orderBook.Bids.Length() > 0 && price.LessThanOrEqual(maxPrice) {
			bestPriceBids := orderBook.Bids.MaxPriceList()
			quantityToTrade, newTrades = orderBook.processOrderList(BID, bestPriceBids, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
			maxPrice = orderBook.Bids.MaxPrice()
		}

		if quantityToTrade.GreaterThan(decimal.Zero) {
			quote["order_id"] = strconv.Itoa(orderBook.NextOrderID)
			quote["quantity"] = quantityToTrade.String()
			orderBook.Asks.InsertOrder(quote)
			orderInBook = quote
		}
	}
	return trades, orderInBook
}

// ProcessOrder : process the order
func (orderBook *OrderBook) ProcessOrder(quote map[string]string, verbose bool) ([]map[string]string, map[string]string) {
	orderType := quote["type"]
	var orderInBook map[string]string
	var trades []map[string]string

	orderBook.UpdateTime()
	// quote["timestamp"] = strconv.Itoa(orderBook.Time)
	orderBook.NextOrderID++

	if orderType == "market" {
		trades = orderBook.processMarketOrder(quote, verbose)
	} else {
		trades, orderInBook = orderBook.processLimitOrder(quote, verbose)
	}
	return trades, orderInBook
}

// processOrderList : process the order list
func (orderBook *OrderBook) processOrderList(side string, orderList *OrderList, quantityStillToTrade decimal.Decimal, quote map[string]string, verbose bool) (decimal.Decimal, []map[string]string) {
	quantityToTrade := quantityStillToTrade
	var trades []map[string]string

	for orderList.Length > 0 && quantityToTrade.GreaterThan(decimal.Zero) {
		headOrder := orderList.HeadOrder
		tradedPrice := headOrder.Price

		var newBookQuantity decimal.Decimal
		var tradedQuantity decimal.Decimal

		if quantityToTrade.LessThan(headOrder.Quantity) {
			tradedQuantity = quantityToTrade
			// Do the transaction
			newBookQuantity = headOrder.Quantity.Sub(quantityToTrade)
			headOrder.UpdateQuantity(newBookQuantity, headOrder.Timestamp)
			quantityToTrade = decimal.Zero

		} else if quantityToTrade.Equal(headOrder.Quantity) {
			tradedQuantity = quantityToTrade
			if side == BID {
				orderBook.Bids.RemoveOrderByID(headOrder.OrderID)
			} else {
				orderBook.Asks.RemoveOrderByID(headOrder.OrderID)
			}
			quantityToTrade = decimal.Zero

		} else {
			tradedQuantity = headOrder.Quantity
			if side == BID {
				orderBook.Bids.RemoveOrderByID(headOrder.OrderID)
			} else {
				orderBook.Asks.RemoveOrderByID(headOrder.OrderID)
			}
		}

		if verbose {
			fmt.Printf("TRADE: Time - %d, Price - %s, Quantity - %s, TradeID - %s, Matching TradeID - %s\n",
				orderBook.Time, tradedPrice, tradedQuantity, headOrder.TradeID, quote["trade_id"])
		}

		transactionRecord := make(map[string]string)
		transactionRecord["timestamp"] = strconv.Itoa(orderBook.Time)
		transactionRecord["price"] = tradedPrice.String()
		transactionRecord["quantity"] = tradedQuantity.String()
		transactionRecord["time"] = strconv.Itoa(orderBook.Time)

		orderBook.Deque.Append(transactionRecord)
		trades = append(trades, transactionRecord)
	}
	return quantityToTrade, trades
}

// CancelOrder : cancel the order
func (orderBook *OrderBook) CancelOrder(side string, orderID int) {
	orderBook.UpdateTime()

	if side == BID {
		if orderBook.Bids.OrderExist(strconv.Itoa(orderID)) {
			orderBook.Bids.RemoveOrderByID(strconv.Itoa(orderID))
		}
	} else {
		if orderBook.Asks.OrderExist(strconv.Itoa(orderID)) {
			orderBook.Asks.RemoveOrderByID(strconv.Itoa(orderID))
		}
	}
}

// ModifyOrder : modify the order
func (orderBook *OrderBook) ModifyOrder(quoteUpdate map[string]string, orderID int) {
	orderBook.UpdateTime()

	side := quoteUpdate["side"]
	quoteUpdate["order_id"] = strconv.Itoa(orderID)
	quoteUpdate["timestamp"] = strconv.Itoa(orderBook.Time)

	if side == BID {
		if orderBook.Bids.OrderExist(strconv.Itoa(orderID)) {
			orderBook.Bids.UpdateOrder(quoteUpdate)
		}
	} else {
		if orderBook.Asks.OrderExist(strconv.Itoa(orderID)) {
			orderBook.Asks.UpdateOrder(quoteUpdate)
		}
	}
}

// VolumeAtPrice : get volume at the current price
func (orderBook *OrderBook) VolumeAtPrice(side string, price decimal.Decimal) decimal.Decimal {
	if side == BID {
		volume := decimal.Zero
		if orderBook.Bids.PriceExist(price) {
			volume = orderBook.Bids.PriceList(price).Volume
		}
		return volume

	} else {
		volume := decimal.Zero
		if orderBook.Asks.PriceExist(price) {
			volume = orderBook.Asks.PriceList(price).Volume
		}
		return volume
	}
}
