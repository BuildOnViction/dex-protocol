package orderbook

import (
	"bytes"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tomochain/backend-matching-engine/utils/math"
)

// Order : info that will be store in ipfs
type OrderItem struct {
	Timestamp uint64   `json:"timestamp"`
	Quantity  *big.Int `json:"quantity"`
	Price     *big.Int `json:"price"`
	// OrderID   string          `json:"orderID"`
	TradeID string `json:"tradeID"`
	// these following fields can lead to recursive problem
	// NextOrder *Order     `json:"-"`
	// PrevOrder *Order     `json:"-"`
	// OrderList *OrderList `json:"-"`

	NextOrder []byte `json:"-"`
	PrevOrder []byte `json:"-"`
	OrderList []byte `json:"-"`
}

type Order struct {
	Item *OrderItem
	Key  []byte `json:"orderID"`
}

func (o *Order) String() string {
	return ToJSON(o)
}

func (o *Order) GetNextOrder(orderList *OrderList) *Order {
	nextOrder := orderList.GetOrder(o.Item.NextOrder)

	return nextOrder
}

func (o *Order) GetPrevOrder(orderList *OrderList) *Order {
	prevOrder := orderList.GetOrder(o.Item.PrevOrder)

	return prevOrder
}

// NewOrder : create new order with quote ( can be ethereum address )
func NewOrder(quote map[string]string, orderList []byte) *Order {
	timestamp, _ := strconv.ParseUint(quote["timestamp"], 10, 64)
	quantity, _ := new(big.Int).SetString(quote["quantity"], 10)
	price, _ := new(big.Int).SetString(quote["price"], 10)
	orderID := quote["order_id"]
	tradeID := quote["trade_id"]
	orderItem := &OrderItem{
		Timestamp: timestamp,
		Quantity:  quantity,
		Price:     price,
		// OrderID:   orderID,
		TradeID:   tradeID,
		NextOrder: nil,
		PrevOrder: nil,
		OrderList: orderList,
	}

	// key should be Hash for compatible with smart contract
	order := &Order{
		Key:  common.StringToHash(orderID).Bytes(),
		Item: orderItem,
	}

	return order
}

// UpdateQuantity : update quantity of the order
func (o *Order) UpdateQuantity(orderList *OrderList, newQuantity *big.Int, newTimestamp uint64) {
	if newQuantity.Cmp(o.Item.Quantity) > 0 && !bytes.Equal(orderList.Item.TailOrder, o.Key) {
		orderList.MoveToTail(o)
	}
	// update volume and modified timestamp
	orderList.Item.Volume = math.Sub(orderList.Item.Volume, math.Sub(o.Item.Quantity, newQuantity))
	o.Item.Timestamp = newTimestamp
	o.Item.Quantity = newQuantity

	orderList.Save()
}
