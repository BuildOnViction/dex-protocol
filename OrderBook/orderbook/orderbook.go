package orderbook

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// Ask : ask constant
	Ask = "ask"
	// Bid : bid constant
	Bid    = "bid"
	Market = "market"
	Limit  = "limit"

	// we use a big number as segment for storing order, order list from order tree slot.
	// as sequential id
	SlotSegment = common.AddressLength
)

type OrderBookItem struct {
	Timestamp     uint64 `json:"time"`
	NextOrderID   uint64 `json:"nextOrderID"`
	MaxPricePoint uint64 `json:"maxVolume"` // maximum
	Name          string `json:"name"`
}

// OrderBook : list of orders
type OrderBook struct {
	db   *BatchDatabase // this is for orderBook
	Bids *OrderTree     `json:"bids"`
	Asks *OrderTree     `json:"asks"`
	Item *OrderBookItem

	Key  []byte
	slot *big.Int
}

// NewOrderBook : return new order book
func NewOrderBook(name string, db *BatchDatabase) *OrderBook {

	// we can implement using only one DB to faciliate cache engine
	// so that we use a big.Int number to seperate domain of the keys
	// like this keccak("orderBook") + key
	// orderBookPath := path.Join(datadir, "orderbook")
	// db := NewBatchDatabase(orderBookPath, 0, 0)

	item := &OrderBookItem{
		NextOrderID: 0,
		Name:        strings.ToLower(name),
	}

	// do slot with hash to prevent collision

	// we convert to lower case, so even with name as contract address, it is still correct
	// without converting back from hex to bytes
	key := crypto.Keccak256([]byte(item.Name))
	slot := new(big.Int).SetBytes(key)

	// hash ( orderBookKey . orderTreeKey )
	// bidsKey := crypto.Keccak256(key, bidSlotKey)
	// asksKey := crypto.Keccak256(key, askSlotKey)

	// we just increase the segment at the most byte at address length level to avoid conflict
	// somehow it is like 2 hashes has the same common prefix and it is very difficult to resolve
	// the order id start at orderbook slot
	// the price of order tree start at order tree slot
	bidsKey := GetSegmentHash(key, 1, SlotSegment)
	asksKey := GetSegmentHash(key, 2, SlotSegment)

	orderBook := &OrderBook{
		db:   db,
		Item: item,
		slot: slot,
		Key:  key,
	}

	bids := NewOrderTree(db, bidsKey, orderBook)
	asks := NewOrderTree(db, asksKey, orderBook)

	// set asks and bids
	orderBook.Bids = bids
	orderBook.Asks = asks
	// orderBook.Restore()

	// no need to update when there is no operation yet
	orderBook.UpdateTime()

	return orderBook
}

func (orderBook *OrderBook) SetDebug(debug bool) {
	orderBook.db.Debug = debug
}

func (orderBook *OrderBook) Save() error {

	orderBook.Asks.Save()
	orderBook.Bids.Save()

	// orderBookBytes, _ := rlp.EncodeToBytes(orderBook.Item)

	// batch.Put([]byte("asks"), asksBytes)
	// batch.Put([]byte("bids"), bidsBytes)
	// batch.Put([]byte("orderBook"), orderBookBytes)

	// commit
	// return batch.Write()
	return orderBook.db.Put(orderBook.Key, orderBook.Item)
}

// commit everything by trigger db.Commit, later we can map custom encode and decode based on item
func (orderBook *OrderBook) Commit() error {
	return orderBook.db.Commit()
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

	val, err := orderBook.db.Get(orderBook.Key, orderBook.Item)
	if err == nil {
		// 	// return rlp.DecodeBytes(orderBookBytes, orderBook.Item)
		orderBook.Item = val.(*OrderBookItem)
	}

	return err
}

func (orderBook *OrderBook) GetOrderIDFromBook(key []byte) uint64 {
	orderSlot := new(big.Int).SetBytes(key)
	return Sub(orderSlot, orderBook.slot).Uint64()
}

func (orderBook *OrderBook) GetOrderIDFromKey(key []byte) []byte {
	orderSlot := new(big.Int).SetBytes(key)
	// fmt.Println("FAIL", key, orderList.slot)
	return common.BigToHash(Add(orderBook.slot, orderSlot)).Bytes()
}

func (orderBook *OrderBook) GetOrder(key []byte) *Order {
	if orderBook.db.IsEmptyKey(key) {
		return nil
	}
	// orderID := key
	storedKey := orderBook.GetOrderIDFromKey(key)
	orderItem := &OrderItem{}
	// var orderItem *OrderItem
	val, err := orderBook.db.Get(storedKey, orderItem)
	if err != nil {
		fmt.Printf("Key not found :%x, %v\n", storedKey, err)
		return nil
	}

	order := &Order{
		Item: val.(*OrderItem),
		Key:  key,
	}
	return order
}

// we need to store orderBook information as well
// Volume    *big.Int `json:"volume"`    // Contains total quantity from all Orders in tree
// 	NumOrders int             `json:"numOrders"` // Contains count of Orders in tree
// 	Depth

func (orderBook *OrderBook) String(startDepth int) string {
	tabs := strings.Repeat("\t", startDepth)
	return fmt.Sprintf("%s{\n\t%sName: %s\n\t%sTimestamp: %d\n\t%sNextOrderID: %d\n\t%sBids: %s\n\t%sAsks: %s\n%s}\n",
		tabs,
		tabs, orderBook.Item.Name, tabs, orderBook.Item.Timestamp, tabs, orderBook.Item.NextOrderID,
		tabs, orderBook.Bids.String(startDepth+1), tabs, orderBook.Asks.String(startDepth+1),
		tabs)
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
	// speedup the comparison, do not assign because it is pointer
	zero := Zero()
	if side == Bid {
		for quantityToTrade.Cmp(zero) > 0 && orderBook.Asks.NotEmpty() {
			bestPriceAsks := orderBook.Asks.MinPriceList()
			quantityToTrade, newTrades = orderBook.processOrderList(Ask, bestPriceAsks, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
		}
		// } else if side == Ask {
	} else {
		for quantityToTrade.Cmp(zero) > 0 && orderBook.Bids.NotEmpty() {
			bestPriceBids := orderBook.Bids.MaxPriceList()
			quantityToTrade, newTrades = orderBook.processOrderList(Bid, bestPriceBids, quantityToTrade, quote, verbose)
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
	// speedup the comparison, do not assign because it is pointer
	zero := Zero()

	if side == Bid {
		minPrice := orderBook.Asks.MinPrice()
		for quantityToTrade.Cmp(zero) > 0 && orderBook.Asks.NotEmpty() && price.Cmp(minPrice) >= 0 {
			bestPriceAsks := orderBook.Asks.MinPriceList()
			quantityToTrade, newTrades = orderBook.processOrderList(Ask, bestPriceAsks, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
			minPrice = orderBook.Asks.MinPrice()
		}

		if quantityToTrade.Cmp(zero) > 0 {
			quote["order_id"] = strconv.FormatUint(orderBook.Item.NextOrderID, 10)
			quote["quantity"] = quantityToTrade.String()
			orderBook.Bids.InsertOrder(quote)
			orderInBook = quote
		}

		// } else if side == Ask {
	} else {
		maxPrice := orderBook.Bids.MaxPrice()
		for quantityToTrade.Cmp(zero) > 0 && orderBook.Bids.NotEmpty() && price.Cmp(maxPrice) <= 0 {
			bestPriceBids := orderBook.Bids.MaxPriceList()
			quantityToTrade, newTrades = orderBook.processOrderList(Bid, bestPriceBids, quantityToTrade, quote, verbose)
			trades = append(trades, newTrades...)
			maxPrice = orderBook.Bids.MaxPrice()
		}

		if quantityToTrade.Cmp(zero) > 0 {
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

	if orderType == Market {
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
	// speedup the comparison, do not assign because it is pointer
	zero := Zero()
	// var watchDog = 0
	// fmt.Printf("CMP problem :%t - %t\n", quantityToTrade.Cmp(Zero()) > 0, IsGreaterThan(quantityToTrade, Zero()))
	for orderList.Item.Length > 0 && quantityToTrade.Cmp(zero) > 0 {

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
			if side == Bid {
				// orderBook.Bids.RemoveOrderByID(headOrder.Key)
				orderBook.Bids.RemoveOrder(headOrder)
			} else {
				// orderBook.Asks.RemoveOrderByID(headOrder.Key)
				orderBook.Asks.RemoveOrder(headOrder)
			}
			quantityToTrade = Zero()

		} else {
			tradedQuantity = CloneBigInt(headOrder.Item.Quantity)
			if side == Bid {
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

// CancelOrder : cancel the order, just need ID, side and price, of course order must belong
// to a price point as well
func (orderBook *OrderBook) CancelOrder(side string, orderID uint64, price *big.Int) error {
	orderBook.UpdateTime()
	key := GetKeyFromBig(big.NewInt(int64(orderID)))
	var err error
	if side == Bid {
		order := orderBook.Bids.GetOrder(key, price)
		if order != nil {
			_, err = orderBook.Bids.RemoveOrder(order)
		}
		// if orderBook.Bids.OrderExist(key, price) {
		// 	orderBook.Bids.RemoveOrder(order)
		// }
	} else {

		order := orderBook.Asks.GetOrder(key, price)
		if order != nil {
			_, err = orderBook.Asks.RemoveOrder(order)
		}

		// if orderBook.Asks.OrderExist(key) {
		// 	orderBook.Asks.RemoveOrder(order)
		// }
	}

	return err
}

func (orderBook *OrderBook) UpdateOrder(quoteUpdate map[string]string) error {
	orderID, err := strconv.ParseUint(quoteUpdate["order_id"], 10, 64)
	if err == nil {
		price, ok := new(big.Int).SetString(quoteUpdate["price"], 10)
		if !ok {
			return fmt.Errorf("Price is not correct :%s", quoteUpdate["price"])
		}

		return orderBook.ModifyOrder(quoteUpdate, orderID, price)
	}
	return err
}

// ModifyOrder : modify the order
func (orderBook *OrderBook) ModifyOrder(quoteUpdate map[string]string, orderID uint64, price *big.Int) error {
	orderBook.UpdateTime()

	side := quoteUpdate["side"]
	quoteUpdate["order_id"] = strconv.FormatUint(orderID, 10)
	quoteUpdate["timestamp"] = strconv.FormatUint(orderBook.Item.Timestamp, 10)
	key := GetKeyFromBig(ToBigInt(quoteUpdate["order_id"]))
	if side == Bid {

		if orderBook.Bids.OrderExist(key, price) {
			return orderBook.Bids.UpdateOrder(quoteUpdate)
		}
		// if orderBook.Bids.OrderExist(key) {
		// 	orderBook.Bids.UpdateOrder(quoteUpdate)
		// }
	} else {

		if orderBook.Asks.OrderExist(key, price) {
			return orderBook.Asks.UpdateOrder(quoteUpdate)
		}
	}

	return nil
}

// VolumeAtPrice : get volume at the current price
func (orderBook *OrderBook) VolumeAtPrice(side string, price *big.Int) *big.Int {
	volume := Zero()
	if side == Bid {
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
