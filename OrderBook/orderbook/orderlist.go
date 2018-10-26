package orderbook

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// Item : comparable
type Item interface {
	Less(than Item) bool
}

const (
	// LimitDepthPrint : the maximum depth of order list to be printed
	LimitDepthPrint = 20
)

// OrderList : order list
type OrderList struct {
	HeadOrder *Order          `json:"headOrder"`
	TailOrder *Order          `json:"tailOrder"`
	Length    int             `json:"length"`
	Volume    decimal.Decimal `json:"volume"`
	Price     decimal.Decimal `json:"price"`
}

// NewOrderList : return new OrderList
func NewOrderList(price decimal.Decimal) *OrderList {
	return &OrderList{
		HeadOrder: nil,
		TailOrder: nil,
		Length:    0,
		Volume:    decimal.Zero,
		Price:     price,
	}
}

func (orderList *OrderList) String(startDepth int) string {
	var buffer bytes.Buffer
	tabs := strings.Repeat("\t", startDepth)
	buffer.WriteString(fmt.Sprintf("{\n\t%sLength: %d\n\t%sVolume: %v\n\t%sPrice: %v",
		tabs, orderList.Length, tabs, orderList.Volume, tabs, orderList.Price))

	buffer.WriteString("\n\t")
	buffer.WriteString(tabs)
	buffer.WriteString("Head:")
	linkedList := orderList.HeadOrder
	depth := 0
	for linkedList != nil {
		depth++
		spaces := strings.Repeat(" ", depth)
		if depth > LimitDepthPrint {
			buffer.WriteString(fmt.Sprintf("\n\t%s%s |-> %s %d left", tabs, spaces, "...", orderList.Length-LimitDepthPrint))
			break
		}
		buffer.WriteString(fmt.Sprintf("\n\t%s%s |-> %s", tabs, spaces, linkedList.String()))
		linkedList = linkedList.NextOrder
	}
	buffer.WriteString("\n\t")
	buffer.WriteString(tabs)
	buffer.WriteString("Tail:")
	linkedList = orderList.TailOrder
	depth = 0
	for linkedList != nil {
		depth++
		spaces := strings.Repeat(" ", depth)
		if depth > LimitDepthPrint {
			buffer.WriteString(fmt.Sprintf("\n\t%s%s <-| %s %d left", tabs, spaces, "...", orderList.Length-LimitDepthPrint))
			break
		}
		buffer.WriteString(fmt.Sprintf("\n\t%s%s <-| %s", tabs, spaces, linkedList.String()))
		linkedList = linkedList.PrevOrder
	}
	buffer.WriteString("\n")
	buffer.WriteString(tabs)
	buffer.WriteString("}")
	return buffer.String()
}

// Less : compare if this order list is less than compared object
func (orderList *OrderList) Less(than Item) bool {
	// cast to OrderList pointer
	return orderList.Price.LessThan(than.(*OrderList).Price)
}

// AppendOrder : append order into the order list
func (orderList *OrderList) AppendOrder(order *Order) {
	if orderList.Length == 0 {
		order.NextOrder = nil
		order.PrevOrder = nil
		orderList.HeadOrder = order
		orderList.TailOrder = order
	} else {
		order.PrevOrder = orderList.TailOrder
		order.NextOrder = nil
		orderList.TailOrder.NextOrder = order
		orderList.TailOrder = order
	}
	orderList.Length++
	orderList.Volume = orderList.Volume.Add(order.Quantity)
}

// RemoveOrder : remove order from the order list
func (orderList *OrderList) RemoveOrder(order *Order) {
	orderList.Volume = orderList.Volume.Sub(order.Quantity)
	orderList.Length--
	if orderList.Length == 0 {
		return
	}

	nextOrder := order.NextOrder
	prevOrder := order.PrevOrder

	if nextOrder != nil && prevOrder != nil {
		nextOrder.PrevOrder = prevOrder
		prevOrder.NextOrder = nextOrder
	} else if nextOrder != nil {
		nextOrder.PrevOrder = nil
		orderList.HeadOrder = nextOrder
	} else if prevOrder != nil {
		prevOrder.NextOrder = nil
		orderList.TailOrder = prevOrder
	}
}

// MoveToTail : move order to the end of the order list
func (orderList *OrderList) MoveToTail(order *Order) {
	if order.PrevOrder != nil { // This Order is not the first Order in the OrderList
		order.PrevOrder.NextOrder = order.NextOrder // Link the previous Order to the next Order, then move the Order to tail
	} else { // This Order is the first Order in the OrderList
		orderList.HeadOrder = order.NextOrder // Make next order the first
	}
	order.NextOrder.PrevOrder = order.PrevOrder

	// Move Order to the last position. Link up the previous last position Order.
	orderList.TailOrder.NextOrder = order
	orderList.TailOrder = order
}

// String : travel the list to print it in nice format
