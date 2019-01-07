package orderbook

import (
	"fmt"
	"math/big"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	// ASK : ask constant
	ASK = "ask"
	// BID : bid constant
	BID              = "bid"
	ORDERTYPE_MARKET = "market"
	ORDERTYPE_LIMIT  = "limit"
)

type OrderBookItem struct {
	Timestamp   uint64 `json:"time"`
	NextOrderID uint64 `json:"nextOrderID"`
}

// OrderBook : list of orders
type OrderBook struct {
	db   *ethdb.LDBDatabase // this is for orderBook
	Bids *OrderTree         `json:"bids"`
	Asks *OrderTree         `json:"asks"`
	Item *OrderBookItem
}

// NewOrderBook : return new order book
func NewOrderBook(datadir string) *OrderBook {

	// we can implement using only one DB to faciliate cache engine
	// so that we use a big.Int number to seperate domain of the keys
	// like this keccak("orderBook") + key
	orderBookPath := path.Join(datadir, "orderBook")
	bidsPath := path.Join(datadir, "bids")
	asksPath := path.Join(datadir, "asks")
	bids := NewOrderTree(bidsPath)
	asks := NewOrderTree(asksPath)

	db, _ := ethdb.NewLDBDatabase(orderBookPath, 0, 0)

	item := &OrderBookItem{
		NextOrderID: 0,
	}

	orderBook := &OrderBook{
		db:   db,
		Bids: bids,
		Asks: asks,
		Item: item,
	}

	orderBook.Restore()

	// no need to update when there is no operation yet
	orderBook.UpdateTime()

	return orderBook
}

func (orderBook *OrderBook) Save() error {

	orderBook.Asks.Save()
	orderBook.Bids.Save()

	orderBookBytes, _ := rlp.EncodeToBytes(orderBook.Item)

	// batch.Put([]byte("asks"), asksBytes)
	// batch.Put([]byte("bids"), bidsBytes)
	// batch.Put([]byte("orderBook"), orderBookBytes)

	// commit
	// return batch.Write()
	return orderBook.db.Put([]byte("orderBook"), orderBookBytes)
}

func (orderBook *OrderBook) Restore() error {

	// if asksBytes, err := orderBook.db.Get([]byte("asks")); err != nil {
	// 	rlp.DecodeBytes(asksBytes, orderBook.Asks.Item)
	// }
	// if bidsBytes, err := orderBook.db.Get([]byte("bids")); err != nil {
	// 	rlp.DecodeBytes(bidsBytes, orderBook.Bids.Item)
	// }

	orderBook.Asks.Restore()
	orderBook.Bids.Restore()

	orderBookBytes, err := orderBook.db.Get([]byte("orderBook"))
	if err == nil {
		return rlp.DecodeBytes(orderBookBytes, orderBook.Item)
	}
	return err
}

// we need to store orderBook information as well
// Volume    *big.Int `json:"volume"`    // Contains total quantity from all Orders in tree
// 	NumOrders int             `json:"numOrders"` // Contains count of Orders in tree
// 	Depth

func (orderBook *OrderBook) String(startDepth int) string {
	tabs := strings.Repeat("\t", startDepth)
	return fmt.Sprintf("{\n\t%sBids: %s\n\t%sAsks: %s\n\t%sTimestamp: %d\n\t%sNextOrderID: %d\n%s}\n",
		tabs, orderBook.Bids.String(startDepth+1), tabs, orderBook.Asks.String(startDepth+1), tabs,
		orderBook.Item.Timestamp, tabs, orderBook.Item.NextOrderID, tabs)
}

// UpdateTime : update time for order book
func (orderBook *OrderBook) UpdateTime() {
	timestamp := uint64(time.Now().Unix())
	orderBook.Item.Timestamp = timestamp
}

// BestBid : get the best bid of the order book
func (orderBook *OrderBook) BestBid() (value *big.Int) {
	return orderBook.Bids.MaxPrice()
}

// BestAsk : get the best ask of the order book
func (orderBook *OrderBook) BestAsk() (value *big.Int) {
	return orderBook.Asks.MinPrice()
}

// WorstBid : get the worst bid of the order book
func (orderBook *OrderBook) WorstBid() (value *big.Int) {
	return orderBook.Bids.MinPrice()
}

// WorstAsk : get the worst ask of the order book
func (orderBook *OrderBook) WorstAsk() (value *big.Int) {
	return orderBook.Asks.MaxPrice()
}

// processMarketOrder : process the market order
func (orderBook *OrderBook) processMarketOrder(quote map[string]string, verbose bool) []map[string]string {
	var trades []map[string]string
	quantityToTrade := ToBigInt(quote["quantity"])
	side := quote["side"]
	var newTrades []map[string]string

	if side == BID {
		for quantityToTrade.Cmp(Zero()) > 0 && orderBook.Asks.NotEmpty() {
			bestPriceAsks := orderBook.Asks.MinPriceList()
			quantityToTrade, newTrades = orderBook.processOrderList(ASK, bestPriceAsks, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
		}
		// } else if side == ASK {
	} else {
		for quantityToTrade.Cmp(Zero()) > 0 && orderBook.Bids.NotEmpty() {
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
	quantityToTrade := ToBigInt(quote["quantity"])
	side := quote["side"]
	price := ToBigInt(quote["price"])

	var newTrades []map[string]string
	var orderInBook map[string]string

	if side == BID {
		minPrice := orderBook.Asks.MinPrice()
		for quantityToTrade.Cmp(Zero()) > 0 && orderBook.Asks.NotEmpty() && price.Cmp(minPrice) >= 0 {
			bestPriceAsks := orderBook.Asks.MinPriceList()
			quantityToTrade, newTrades = orderBook.processOrderList(ASK, bestPriceAsks, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
			minPrice = orderBook.Asks.MinPrice()
		}

		if quantityToTrade.Cmp(Zero()) > 0 {
			quote["order_id"] = strconv.FormatUint(orderBook.Item.NextOrderID, 10)
			quote["quantity"] = quantityToTrade.String()
			orderBook.Bids.InsertOrder(quote)
			orderInBook = quote
		}

		// } else if side == ASK {
	} else {
		maxPrice := orderBook.Bids.MaxPrice()
		for quantityToTrade.Cmp(Zero()) > 0 && orderBook.Bids.NotEmpty() && price.Cmp(maxPrice) <= 0 {
			bestPriceBids := orderBook.Bids.MaxPriceList()
			quantityToTrade, newTrades = orderBook.processOrderList(BID, bestPriceBids, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
			maxPrice = orderBook.Bids.MaxPrice()
		}

		if quantityToTrade.Cmp(Zero()) > 0 {
			quote["order_id"] = strconv.FormatUint(orderBook.Item.NextOrderID, 10)
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
	// if we do not use auto-increment orderid, we must set price slot to avoid conflict
	orderBook.Item.NextOrderID++

	if orderType == ORDERTYPE_MARKET {
		trades = orderBook.processMarketOrder(quote, verbose)
	} else {
		trades, orderInBook = orderBook.processLimitOrder(quote, verbose)
	}

	// update orderBook
	orderBook.Save()

	return trades, orderInBook
}

// processOrderList : process the order list
func (orderBook *OrderBook) processOrderList(side string, orderList *OrderList, quantityStillToTrade *big.Int, quote map[string]string, verbose bool) (*big.Int, []map[string]string) {
	quantityToTrade := CloneBigInt(quantityStillToTrade)
	// quantityToTrade := quantityStillToTrade
	var trades []map[string]string
	// var watchDog = 0
	// fmt.Printf("CMP problem :%t - %t\n", quantityToTrade.Cmp(Zero()) > 0, IsGreaterThan(quantityToTrade, Zero()))
	for orderList.Item.Length > 0 && IsStrictlyGreaterThan(quantityToTrade, Zero()) {

		headOrder := orderList.GetOrder(orderList.Item.HeadOrder)
		// fmt.Printf("Head :%s ,%s\n", new(big.Int).SetBytes(orderList.Item.HeadOrder), orderBook.Asks.MinPriceList().String(0))
		if headOrder == nil {
			panic("headOrder is null")
			// return Zero(), trades
		}

		tradedPrice := CloneBigInt(headOrder.Item.Price)

		var newBookQuantity *big.Int
		var tradedQuantity *big.Int

		if IsStrictlySmallerThan(quantityToTrade, headOrder.Item.Quantity) {
			tradedQuantity = CloneBigInt(quantityToTrade)
			// Do the transaction
			newBookQuantity = Sub(headOrder.Item.Quantity, quantityToTrade)
			headOrder.UpdateQuantity(orderList, newBookQuantity, headOrder.Item.Timestamp)
			quantityToTrade = Zero()

		} else if IsEqual(quantityToTrade, headOrder.Item.Quantity) {
			tradedQuantity = CloneBigInt(quantityToTrade)
			if side == BID {
				// orderBook.Bids.RemoveOrderByID(headOrder.Key)
				orderBook.Bids.RemoveOrder(headOrder)
			} else {
				// orderBook.Asks.RemoveOrderByID(headOrder.Key)
				orderBook.Asks.RemoveOrder(headOrder)
			}
			quantityToTrade = Zero()

		} else {
			tradedQuantity = CloneBigInt(headOrder.Item.Quantity)
			if side == BID {
				// orderBook.Bids.RemoveOrderByID(headOrder.Key)
				orderBook.Bids.RemoveOrder(headOrder)
			} else {
				// orderBook.Asks.RemoveOrderByID(headOrder.Key)
				// fmt.Printf("\nBEFORE : %s\n\n", orderList.String(0))
				// orderList, _ = orderBook.Asks.RemoveOrder(headOrder)
				orderBook.Asks.RemoveOrderFromOrderList(headOrder, orderList)
				// fmt.Println("AFTER DELETE", orderList.String(0))
				// fmt.Printf("\nAFTER : %x, %s\n\n", orderList.Item.HeadOrder, orderList.String(0))
			}
		}

		if verbose {
			fmt.Printf("TRADE: Timestamp - %d, Price - %s, Quantity - %s, TradeID - %s, Matching TradeID - %s\n",
				orderBook.Item.Timestamp, tradedPrice, tradedQuantity, headOrder.Item.TradeID, quote["trade_id"])
			// fmt.Println(headOrder)
			// watchDog++
			// if watchDog > 10 {
			// panic("stop")
			// }

		}

		transactionRecord := make(map[string]string)
		transactionRecord["timestamp"] = strconv.FormatUint(orderBook.Item.Timestamp, 10)
		transactionRecord["price"] = tradedPrice.String()
		transactionRecord["quantity"] = tradedQuantity.String()

		trades = append(trades, transactionRecord)
	}
	return quantityToTrade, trades
}

// CancelOrder : cancel the order
func (orderBook *OrderBook) CancelOrder(side string, orderID int, price *big.Int) {
	orderBook.UpdateTime()
	key := GetKeyFromBig(big.NewInt(int64(orderID)))

	if side == BID {
		order := orderBook.Bids.GetOrder(key, price)
		if order != nil {
			orderBook.Bids.RemoveOrder(order)
		}
		// if orderBook.Bids.OrderExist(key, price) {
		// 	orderBook.Bids.RemoveOrder(order)
		// }
	} else {

		order := orderBook.Asks.GetOrder(key, price)
		if order != nil {
			orderBook.Asks.RemoveOrder(order)
		}

		// if orderBook.Asks.OrderExist(key) {
		// 	orderBook.Asks.RemoveOrder(order)
		// }
	}
}

// ModifyOrder : modify the order
func (orderBook *OrderBook) ModifyOrder(quoteUpdate map[string]string, orderID int, price *big.Int) {
	orderBook.UpdateTime()

	side := quoteUpdate["side"]
	quoteUpdate["order_id"] = strconv.Itoa(orderID)
	quoteUpdate["timestamp"] = strconv.FormatUint(orderBook.Item.Timestamp, 10)
	key := GetKeyFromBig(ToBigInt(quoteUpdate["order_id"]))
	if side == BID {

		if orderBook.Bids.OrderExist(key, price) {
			orderBook.Bids.UpdateOrder(quoteUpdate)
		}
		// if orderBook.Bids.OrderExist(key) {
		// 	orderBook.Bids.UpdateOrder(quoteUpdate)
		// }
	} else {

		if orderBook.Asks.OrderExist(key, price) {
			orderBook.Asks.UpdateOrder(quoteUpdate)
		}
	}
}

// VolumeAtPrice : get volume at the current price
func (orderBook *OrderBook) VolumeAtPrice(side string, price *big.Int) *big.Int {
	volume := Zero()
	if side == BID {
		if orderBook.Bids.PriceExist(price) {
			orderList := orderBook.Bids.PriceList(price)
			// incase we use cache for PriceList
			volume = CloneBigInt(orderList.Item.Volume)
		}
	} else {
		// other case
		if orderBook.Asks.PriceExist(price) {
			orderList := orderBook.Asks.PriceList(price)
			volume = CloneBigInt(orderList.Item.Volume)
		}
	}

	return volume

}
