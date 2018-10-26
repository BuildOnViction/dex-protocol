package orderbook

import (
	"strconv"

	"github.com/shopspring/decimal"
)

// Order : info that will be store in ipfs
type Order struct {
	Timestamp int             `json:"timestamp"`
	Quantity  decimal.Decimal `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
	OrderID   string          `json:"orderID"`
	TradeID   string          `json:"tradeID"`
	// these following fields can lead to recursive problem
	NextOrder *Order     `json:"-"`
	PrevOrder *Order     `json:"-"`
	OrderList *OrderList `json:"-"`
}

func (o *Order) String() string {
	return ToJSON(o)
}

// NewOrder : create new order with quote ( can be ethereum address )
func NewOrder(quote map[string]string, orderList *OrderList) *Order {
	timestamp, _ := strconv.Atoi(quote["timestamp"])
	quantity, _ := decimal.NewFromString(quote["quantity"])
	price, _ := decimal.NewFromString(quote["price"])
	orderID := quote["order_id"]
	tradeID := quote["trade_id"]
	return &Order{
		Timestamp: timestamp,
		Quantity:  quantity,
		Price:     price,
		OrderID:   orderID,
		TradeID:   tradeID,
		NextOrder: nil,
		PrevOrder: nil,
		OrderList: orderList,
	}
}

// UpdateQuantity : update quantity of the order
func (o *Order) UpdateQuantity(newQuantity decimal.Decimal, newTimestamp int) {
	if newQuantity.GreaterThan(o.Quantity) && o.OrderList.TailOrder != o {
		o.OrderList.MoveToTail(o)
	}
	// update volume and modified timestamp
	o.OrderList.Volume = o.OrderList.Volume.Sub(o.Quantity.Sub(newQuantity))
	o.Timestamp = newTimestamp
	o.Quantity = newQuantity
}
