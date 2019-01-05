package orderbook

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tomochain/backend-matching-engine/utils/math"
)

// Item : comparable
// type Item interface {
// 	Less(than Item) bool
// }

const (
	// LimitDepthPrint : the maximum depth of order list to be printed
	LimitDepthPrint = 20
)

var emptyKey = []byte{}

type OrderListItem struct {
	// HeadOrder *Order          `json:"headOrder"`
	// TailOrder *Order          `json:"tailOrder"`
	HeadOrder []byte   `json:"headOrder"`
	TailOrder []byte   `json:"tailOrder"`
	Length    uint64   `json:"length"`
	Volume    *big.Int `json:"volume"`
	Price     *big.Int `json:"price"`
}

// OrderList : order list
type OrderList struct {
	// db      *ethdb.LDBDatabase
	orderTree *OrderTree
	// orderDB   *ethdb.LDBDatabase
	Item *OrderListItem
	Key  []byte
}

// NewOrderList : return new OrderList
func NewOrderList(price *big.Int, orderTree *OrderTree) *OrderList {
	item := &OrderListItem{
		HeadOrder: nil,
		TailOrder: nil,
		Length:    0,
		Volume:    big.NewInt(0),
		Price:     price,
	}

	key := orderTree.getKeyFromPrice(price)

	return &OrderList{
		Item:      item,
		Key:       key,
		orderTree: orderTree,
	}
}

func (orderList *OrderList) HeadOrderKey(keys ...[]byte) []byte {
	if len(keys) == 1 {
		orderList.Item.HeadOrder = keys[0]
	}

	return orderList.Item.HeadOrder
}

func (orderList *OrderList) GetOrder(key []byte) *Order {
	if orderList.isEmptyKey(key) {
		return nil
	}
	return orderList.orderTree.Order(key)
}

func (orderList *OrderList) isEmptyKey(key []byte) bool {
	return key == nil || len(key) == 0 || bytes.Equal(key, emptyKey)
}

func (orderList *OrderList) String(startDepth int) string {
	var buffer bytes.Buffer
	tabs := strings.Repeat("\t", startDepth)
	buffer.WriteString(fmt.Sprintf("{\n\t%sLength: %d\n\t%sVolume: %v\n\t%sPrice: %v",
		tabs, orderList.Item.Length, tabs, orderList.Item.Volume, tabs, orderList.Item.Price))

	buffer.WriteString("\n\t")
	buffer.WriteString(tabs)
	buffer.WriteString("Head:")
	linkedList := orderList.GetOrder(orderList.Item.HeadOrder)
	depth := 0
	for linkedList != nil {
		depth++
		spaces := strings.Repeat(" ", depth)
		if depth > LimitDepthPrint {
			buffer.WriteString(fmt.Sprintf("\n\t%s%s |-> %s %d left", tabs, spaces, "...",
				orderList.Item.Length-LimitDepthPrint))
			break
		}
		buffer.WriteString(fmt.Sprintf("\n\t%s%s |-> %s", tabs, spaces, linkedList.String()))
		linkedList = orderList.GetOrder(linkedList.Item.NextOrder)
	}
	buffer.WriteString("\n\t")
	buffer.WriteString(tabs)
	buffer.WriteString("Tail:")
	linkedList = orderList.GetOrder(orderList.Item.TailOrder)
	depth = 0
	for linkedList != nil {
		depth++
		spaces := strings.Repeat(" ", depth)
		if depth > LimitDepthPrint {
			buffer.WriteString(fmt.Sprintf("\n\t%s%s <-| %s %d left", tabs, spaces, "...",
				orderList.Item.Length-LimitDepthPrint))
			break
		}
		buffer.WriteString(fmt.Sprintf("\n\t%s%s <-| %s", tabs, spaces, linkedList.String()))
		linkedList = orderList.GetOrder(linkedList.Item.PrevOrder)
	}
	buffer.WriteString("\n")
	buffer.WriteString(tabs)
	buffer.WriteString("}")
	return buffer.String()
}

// Less : compare if this order list is less than compared object
func (orderList *OrderList) Less(than *OrderList) bool {
	// cast to OrderList pointer
	return orderList.Item.Price.Cmp(than.Item.Price) < 0
}

func (orderList *OrderList) Save() {
	value, err := rlp.EncodeToBytes(orderList.Item)
	if err != nil {
		fmt.Println(err)
		return
	}
	// we use orderlist db file seperated from order
	// orderList.db.Put(orderList.Key, value)
	orderList.orderTree.PriceTree.Put(orderList.Key, value)
	fmt.Printf("Save orderlist key %x, value :%x\n", orderList.Key, value)
}

func (orderList *OrderList) SaveOrder(order *Order) {
	value, err := rlp.EncodeToBytes(order.Item)
	if err != nil {
		fmt.Println(err)
		return
	}

	// using other db to store Order object
	// key := common.BytesToHash(order.Key).Bytes()
	key := order.Key
	orderList.orderTree.OrderDB.Put(key, value)
	fmt.Printf("Save order key : %x, value :%x\n", key, value)
}

// AppendOrder : append order into the order list
func (orderList *OrderList) AppendOrder(order *Order) {
	if orderList.Item.Length == 0 {
		// order.NextOrder = nil
		// order.PrevOrder = nil
		order.Item.NextOrder = emptyKey
		order.Item.PrevOrder = emptyKey

		orderList.Item.HeadOrder = order.Key
		orderList.Item.TailOrder = order.Key

	} else {
		order.Item.PrevOrder = orderList.Item.TailOrder
		order.Item.NextOrder = emptyKey
		tailOrder := orderList.GetOrder(orderList.Item.TailOrder)
		if tailOrder != nil {
			tailOrder.Item.NextOrder = order.Key
			orderList.Item.TailOrder = order.Key
			orderList.SaveOrder(tailOrder)
		}
	}
	orderList.Item.Length++
	orderList.Item.Volume = math.Add(orderList.Item.Volume, order.Item.Quantity)

	orderList.SaveOrder(order)
	orderList.Save()
}

// RemoveOrder : remove order from the order list
func (orderList *OrderList) RemoveOrder(order *Order) {
	orderList.Item.Volume = math.Sub(orderList.Item.Volume, order.Item.Quantity)
	orderList.Item.Length--
	if orderList.Item.Length == 0 {
		return
	}

	nextOrder := orderList.GetOrder(order.Item.NextOrder)
	prevOrder := orderList.GetOrder(order.Item.PrevOrder)

	if nextOrder != nil && prevOrder != nil {
		nextOrder.Item.PrevOrder = prevOrder.Key
		prevOrder.Item.NextOrder = nextOrder.Key

		orderList.SaveOrder(nextOrder)
		orderList.SaveOrder(prevOrder)
	} else if nextOrder != nil {
		nextOrder.Item.PrevOrder = emptyKey
		orderList.Item.HeadOrder = nextOrder.Key

		orderList.SaveOrder(nextOrder)
	} else if prevOrder != nil {
		prevOrder.Item.NextOrder = emptyKey
		orderList.Item.TailOrder = prevOrder.Key

		orderList.SaveOrder(prevOrder)
	}

	orderList.Save()
}

// MoveToTail : move order to the end of the order list
func (orderList *OrderList) MoveToTail(order *Order) {
	if !orderList.isEmptyKey(order.Item.PrevOrder) { // This Order is not the first Order in the OrderList
		prevOrder := orderList.GetOrder(order.Item.PrevOrder)
		if prevOrder != nil {
			prevOrder.Item.NextOrder = order.Item.NextOrder // Link the previous Order to the next Order, then move the Order to tail
			orderList.SaveOrder(prevOrder)
		}

	} else { // This Order is the first Order in the OrderList
		orderList.Item.HeadOrder = order.Item.NextOrder // Make next order the first
	}

	nextOrder := orderList.GetOrder(order.Item.NextOrder)
	if nextOrder != nil {
		nextOrder.Item.PrevOrder = order.Item.PrevOrder
		orderList.SaveOrder(nextOrder)
	}

	// Move Order to the last position. Link up the previous last position Order.
	tailOrder := orderList.GetOrder(orderList.Item.TailOrder)
	if tailOrder != nil {
		tailOrder.Item.NextOrder = order.Key
		orderList.SaveOrder(tailOrder)
	}

	orderList.Item.TailOrder = order.Key
	orderList.Save()
}

// String : travel the list to print it in nice format
