package orderbook

import (
	"bytes"
	"strconv"

	"github.com/shopspring/decimal"
)

// Order : info that will be store in ipfs
type OrderItem struct {
	Timestamp int             `json:"timestamp"`
	Quantity  decimal.Decimal `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
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
	timestamp, _ := strconv.Atoi(quote["timestamp"])
	quantity, _ := decimal.NewFromString(quote["quantity"])
	price, _ := decimal.NewFromString(quote["price"])
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

	order := &Order{
		Key:  []byte(orderID),
		Item: orderItem,
	}

	return order
}

// UpdateQuantity : update quantity of the order
func (o *Order) UpdateQuantity(orderList *OrderList, newQuantity decimal.Decimal, newTimestamp int) {
	if newQuantity.GreaterThan(o.Item.Quantity) && !bytes.Equal(orderList.Item.TailOrder, o.Key) {
		orderList.MoveToTail(o)
	}
	// update volume and modified timestamp
	orderList.Item.Volume = orderList.Item.Volume.Sub(o.Item.Quantity.Sub(newQuantity))
	o.Item.Timestamp = newTimestamp
	o.Item.Quantity = newQuantity

	orderList.Save()
}
