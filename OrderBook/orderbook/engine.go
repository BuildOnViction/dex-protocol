package orderbook

import (
	"fmt"
	"math/big"
	"strings"
)

// Engine : singleton orderbook for testing
type Engine struct {
	Orderbooks map[string]*OrderBook
	db         *BatchDatabase
	// pair and max volume ...
	allowedPairs map[string]*big.Int
}

func NewEngine(datadir string, allowedPairs map[string]*big.Int) *Engine {
	// demo.LogDebug("Creating model", "signerAddress", signer.Address().Hex())
	batchDB := NewBatchDatabaseWithEncode(datadir, 0, 0,
		EncodeBytesItem, DecodeBytesItem)

	fixAllowedPairs := make(map[string]*big.Int)
	for key, value := range allowedPairs {
		fixAllowedPairs[strings.ToLower(key)] = value
	}

	orderbooks := &Engine{
		Orderbooks:   make(map[string]*OrderBook),
		db:           batchDB,
		allowedPairs: fixAllowedPairs,
	}

	return orderbooks
}

func (engine *Engine) GetOrderBook(pairName string) (*OrderBook, error) {
	return engine.getAndCreateIfNotExisted(pairName)
}

func (engine *Engine) hasOrderBook(name string) bool {
	_, ok := engine.Orderbooks[name]
	return ok
}

// commit for all orderbooks
func (engine *Engine) Commit() error {
	return engine.db.Commit()
}

func (engine *Engine) getAndCreateIfNotExisted(pairName string) (*OrderBook, error) {

	name := strings.ToLower(pairName)

	if !engine.hasOrderBook(name) {
		// check allow pair
		if _, ok := engine.allowedPairs[name]; !ok {
			return nil, fmt.Errorf("Orderbook not found for pair :%s", pairName)
		}

		// then create one
		ob := NewOrderBook(name, engine.db)
		if ob != nil {
			ob.Restore()
			engine.Orderbooks[name] = ob
		}
	}

	// return from map
	return engine.Orderbooks[name], nil
}

func (engine *Engine) GetOrder(pairName, orderID string) *Order {
	ob, _ := engine.getAndCreateIfNotExisted(pairName)
	if ob == nil {
		return nil
	}
	key := GetKeyFromString(orderID)
	return ob.GetOrder(key)
}

func (engine *Engine) ProcessOrder(quote map[string]string) ([]map[string]string, map[string]string) {

	ob, _ := engine.getAndCreateIfNotExisted(quote["pairName"])
	var trades []map[string]string
	var orderInBook map[string]string

	if ob != nil {
		// get map as general input, we can set format later to make sure there is no problem
		trades, orderInBook = ob.ProcessOrder(quote, true)
	}

	return trades, orderInBook

}
