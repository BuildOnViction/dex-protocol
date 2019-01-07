package orderbook

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"
)

// OrderItem : info that will be store in database
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
	// *OrderMeta
	NextOrder []byte `json:"-"`
	PrevOrder []byte `json:"-"`
	OrderList []byte `json:"-"`
}

// OrderMeta to help building linked list, there would be a consecutive order meta for each linkedlist
// and slot offset is slot of linkedlist
// type OrderMeta struct {
// 	NextOrder []byte `json:"-"`
// 	PrevOrder []byte `json:"-"`
// 	OrderList []byte `json:"-"`
// }

type Order struct {
	Item *OrderItem
	Key  []byte `json:"orderID"`
}

func (o *Order) String() string {
	return fmt.Sprintf("key : %x, item: %s", o.Key, ToJSON(o.Item))
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
	quantity := ToBigInt(quote["quantity"])
	price := ToBigInt(quote["price"])
	orderID := ToBigInt(quote["order_id"])
	key := GetKeyFromBig(orderID)
	tradeID := quote["trade_id"]
	orderItem := &OrderItem{
		Timestamp: timestamp,
		Quantity:  quantity,
		Price:     price,
		// OrderID:   orderID,
		TradeID:   tradeID,
		NextOrder: emptyKey,
		PrevOrder: emptyKey,
		OrderList: orderList,
	}

	// key should be Hash for compatible with smart contract
	order := &Order{
		Key:  key,
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
	orderList.Item.Volume = Sub(orderList.Item.Volume, Sub(o.Item.Quantity, newQuantity))
	o.Item.Timestamp = newTimestamp
	o.Item.Quantity = CloneBigInt(newQuantity)

	orderList.Save()
}
