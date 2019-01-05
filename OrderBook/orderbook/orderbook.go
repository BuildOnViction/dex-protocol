package orderbook

import (
	"fmt"
	"math/big"
	"path"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tomochain/backend-matching-engine/utils/math"
)

const (
	// ASK : ask constant
	ASK = "ask"
	// BID : bid constant
	BID = "bid"
)

var Zero = big.NewInt(0)

type OrderBookItem struct {
	Time        uint64 `json:"time"`
	NextOrderID uint64 `json:"nextOrderID"`
}

// OrderBook : list of orders
type OrderBook struct {
	db   *ethdb.LDBDatabase // this is for orderbook
	Bids *OrderTree         `json:"bids"`
	Asks *OrderTree         `json:"asks"`
	Item *OrderBookItem
}

// NewOrderBook : return new order book
func NewOrderBook(datadir string) *OrderBook {

	// we can implement using only one DB to faciliate cache engine
	// so that we use a big.Int number to seperate domain of the keys
	// like this keccak("orderbook") + key
	orderbookPath := path.Join(datadir, "orderbook")
	bidsPath := path.Join(datadir, "bids")
	asksPath := path.Join(datadir, "asks")
	bids := NewOrderTree(bidsPath)
	asks := NewOrderTree(asksPath)

	db, _ := ethdb.NewLDBDatabase(orderbookPath, 0, 0)

	item := &OrderBookItem{

		Time:        0,
		NextOrderID: 0,
	}

	orderbook := &OrderBook{
		db:   db,
		Bids: bids,
		Asks: asks,
		Item: item,
	}

	// orderbook.Restore()
	return orderbook
}

func (orderbook *OrderBook) Save() error {

	// commit price tree first
	orderbook.Asks.PriceTree.Commit()
	orderbook.Bids.PriceTree.Commit()

	batch := orderbook.db.NewBatch()

	asksBytes, _ := rlp.EncodeToBytes(orderbook.Asks.Item)
	bidsBytes, _ := rlp.EncodeToBytes(orderbook.Bids.Item)
	orderbookBytes, _ := rlp.EncodeToBytes(orderbook.Item)

	batch.Put([]byte("asks"), asksBytes)
	batch.Put([]byte("bids"), bidsBytes)
	batch.Put([]byte("orderbook"), orderbookBytes)

	// commit
	return batch.Write()
}

func (orderbook *OrderBook) Restore() {

	if asksBytes, err := orderbook.db.Get([]byte("asks")); err != nil {
		rlp.DecodeBytes(asksBytes, orderbook.Asks.Item)
	}
	if bidsBytes, err := orderbook.db.Get([]byte("bids")); err != nil {
		rlp.DecodeBytes(bidsBytes, orderbook.Bids.Item)
	}

	if orderbookBytes, err := orderbook.db.Get([]byte("orderbook")); err != nil {
		rlp.DecodeBytes(orderbookBytes, orderbook.Item)
	}

}

// we need to store orderbook information as well
// Volume    *big.Int `json:"volume"`    // Contains total quantity from all Orders in tree
// 	NumOrders int             `json:"numOrders"` // Contains count of Orders in tree
// 	Depth

func (orderbook *OrderBook) String(startDepth int) string {
	tabs := strings.Repeat("\t", startDepth)
	return fmt.Sprintf("{\n\t%sBids: %s\n\t%sAsks: %s\n\t%sTime: %d\n\t%sNextOrderID: %d\n%s}\n",
		tabs, orderbook.Bids.String(startDepth+1), tabs, orderbook.Asks.String(startDepth+1), tabs,
		orderbook.Item.Time, tabs, orderbook.Item.NextOrderID, tabs)
}

// UpdateTime : update time for order book
func (orderbook *OrderBook) UpdateTime() {
	orderbook.Item.Time++
}

// BestBid : get the best bid of the order book
func (orderbook *OrderBook) BestBid() (value *big.Int) {
	return orderbook.Bids.MaxPrice()
}

// BestAsk : get the best ask of the order book
func (orderbook *OrderBook) BestAsk() (value *big.Int) {
	return orderbook.Asks.MinPrice()
}

// WorstBid : get the worst bid of the order book
func (orderbook *OrderBook) WorstBid() (value *big.Int) {
	return orderbook.Bids.MinPrice()
}

// WorstAsk : get the worst ask of the order book
func (orderbook *OrderBook) WorstAsk() (value *big.Int) {
	return orderbook.Asks.MaxPrice()
}

// processMarketOrder : process the market order
func (orderbook *OrderBook) processMarketOrder(quote map[string]string, verbose bool) []map[string]string {
	var trades []map[string]string
	quantityToTrade, _ := new(big.Int).SetString(quote["quantity"], 10)
	side := quote["side"]
	var newTrades []map[string]string

	if side == BID {
		for quantityToTrade.Cmp(Zero) > 0 && orderbook.Asks.NotEmpty() {
			bestPriceAsks := orderbook.Asks.MinPriceList()
			quantityToTrade, newTrades = orderbook.processOrderList(ASK, bestPriceAsks, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
		}
	} else if side == ASK {
		for quantityToTrade.Cmp(Zero) > 0 && orderbook.Bids.NotEmpty() {
			bestPriceBids := orderbook.Bids.MaxPriceList()
			quantityToTrade, newTrades = orderbook.processOrderList(BID, bestPriceBids, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
		}
	}
	return trades
}

// processLimitOrder : process the limit order, can change the quote
// If not care for performance, we should make a copy of quote to prevent further reference problem
func (orderbook *OrderBook) processLimitOrder(quote map[string]string, verbose bool) ([]map[string]string, map[string]string) {
	var trades []map[string]string
	quantityToTrade, _ := new(big.Int).SetString(quote["quantity"], 10)
	side := quote["side"]
	price, _ := new(big.Int).SetString(quote["price"], 10)
	var newTrades []map[string]string

	var orderInBook map[string]string

	if side == BID {
		minPrice := orderbook.Asks.MinPrice()
		for quantityToTrade.Cmp(Zero) > 0 && orderbook.Asks.NotEmpty() && price.Cmp(minPrice) >= 0 {
			bestPriceAsks := orderbook.Asks.MinPriceList()
			quantityToTrade, newTrades = orderbook.processOrderList(ASK, bestPriceAsks, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
			minPrice = orderbook.Asks.MinPrice()
		}

		if quantityToTrade.Cmp(Zero) > 0 {
			quote["order_id"] = strconv.FormatUint(orderbook.Item.NextOrderID, 10)
			quote["quantity"] = quantityToTrade.String()
			orderbook.Bids.InsertOrder(quote)
			orderInBook = quote
		}

	} else if side == "ask" {
		maxPrice := orderbook.Bids.MaxPrice()
		for quantityToTrade.Cmp(Zero) > 0 && orderbook.Bids.NotEmpty() && price.Cmp(maxPrice) <= 0 {
			bestPriceBids := orderbook.Bids.MaxPriceList()
			quantityToTrade, newTrades = orderbook.processOrderList(BID, bestPriceBids, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
			maxPrice = orderbook.Bids.MaxPrice()
		}

		if quantityToTrade.Cmp(Zero) > 0 {
			quote["order_id"] = strconv.FormatUint(orderbook.Item.NextOrderID, 10)
			quote["quantity"] = quantityToTrade.String()
			orderbook.Asks.InsertOrder(quote)
			orderInBook = quote
		}
	}
	return trades, orderInBook
}

// ProcessOrder : process the order
func (orderbook *OrderBook) ProcessOrder(quote map[string]string, verbose bool) ([]map[string]string, map[string]string) {
	orderType := quote["type"]
	var orderInBook map[string]string
	var trades []map[string]string

	orderbook.UpdateTime()
	// quote["timestamp"] = strconv.Itoa(orderbook.Time)
	orderbook.Item.NextOrderID++

	if orderType == "market" {
		trades = orderbook.processMarketOrder(quote, verbose)
	} else {
		trades, orderInBook = orderbook.processLimitOrder(quote, verbose)
	}
	return trades, orderInBook
}

// processOrderList : process the order list
func (orderbook *OrderBook) processOrderList(side string, orderList *OrderList, quantityStillToTrade *big.Int, quote map[string]string, verbose bool) (*big.Int, []map[string]string) {
	quantityToTrade := quantityStillToTrade
	var trades []map[string]string

	for orderList.Item.Length > 0 && quantityToTrade.Cmp(Zero) > 0 {
		headOrder := orderList.GetOrder(orderList.Item.HeadOrder)
		tradedPrice := headOrder.Item.Price

		var newBookQuantity *big.Int
		var tradedQuantity *big.Int

		if quantityToTrade.Cmp(headOrder.Item.Quantity) < 0 {
			tradedQuantity = quantityToTrade
			// Do the transaction
			newBookQuantity = math.Sub(headOrder.Item.Quantity, quantityToTrade)
			headOrder.UpdateQuantity(orderList, newBookQuantity, headOrder.Item.Timestamp)
			quantityToTrade = Zero

		} else if quantityToTrade.Cmp(headOrder.Item.Quantity) == 0 {
			tradedQuantity = quantityToTrade
			if side == BID {
				orderbook.Bids.RemoveOrderByID(headOrder.Key)
			} else {
				orderbook.Asks.RemoveOrderByID(headOrder.Key)
			}
			quantityToTrade = Zero

		} else {
			tradedQuantity = headOrder.Item.Quantity
			if side == BID {
				orderbook.Bids.RemoveOrderByID(headOrder.Key)
			} else {
				orderbook.Asks.RemoveOrderByID(headOrder.Key)
			}
		}

		if verbose {
			fmt.Printf("TRADE: Time - %d, Price - %s, Quantity - %s, TradeID - %s, Matching TradeID - %s\n",
				orderbook.Item.Time, tradedPrice, tradedQuantity, headOrder.Item.TradeID, quote["trade_id"])
		}

		transactionRecord := make(map[string]string)
		transactionRecord["timestamp"] = strconv.FormatUint(orderbook.Item.Time, 10)
		transactionRecord["price"] = tradedPrice.String()
		transactionRecord["quantity"] = tradedQuantity.String()

		trades = append(trades, transactionRecord)
	}
	return quantityToTrade, trades
}

// CancelOrder : cancel the order
func (orderbook *OrderBook) CancelOrder(side string, orderID int) {
	orderbook.UpdateTime()
	key := []byte(strconv.Itoa(orderID))
	if side == BID {
		if orderbook.Bids.OrderExist(key) {
			orderbook.Bids.RemoveOrderByID(key)
		}
	} else {
		if orderbook.Asks.OrderExist(key) {
			orderbook.Asks.RemoveOrderByID(key)
		}
	}
}

// ModifyOrder : modify the order
func (orderbook *OrderBook) ModifyOrder(quoteUpdate map[string]string, orderID int) {
	orderbook.UpdateTime()

	side := quoteUpdate["side"]
	quoteUpdate["order_id"] = strconv.Itoa(orderID)
	quoteUpdate["timestamp"] = strconv.FormatUint(orderbook.Item.Time, 10)
	key := []byte(quoteUpdate["order_id"])
	if side == BID {
		if orderbook.Bids.OrderExist(key) {
			orderbook.Bids.UpdateOrder(quoteUpdate)
		}
	} else {
		if orderbook.Asks.OrderExist(key) {
			orderbook.Asks.UpdateOrder(quoteUpdate)
		}
	}
}

// VolumeAtPrice : get volume at the current price
func (orderbook *OrderBook) VolumeAtPrice(side string, price *big.Int) *big.Int {
	if side == BID {
		volume := Zero
		if orderbook.Bids.PriceExist(price) {
			orderList := orderbook.Bids.PriceList(price)
			volume = orderList.Item.Volume
		}
		return volume

	}

	// other case
	volume := Zero
	if orderbook.Asks.PriceExist(price) {
		orderList := orderbook.Asks.PriceList(price)
		volume = orderList.Item.Volume
	}
	return volume

}
