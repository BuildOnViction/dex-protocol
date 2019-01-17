package orderbook

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Item : comparable
// type Item interface {
// 	Less(than Item) bool
// }

const (
	// LimitDepthPrint : the maximum depth of order list to be printed
	LimitDepthPrint = 20
)

type OrderListItem struct {
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
	slot      *big.Int
	// orderDB   *ethdb.LDBDatabase
	Item *OrderListItem
	Key  []byte
}

// NewOrderList : return new OrderList
// each orderlist will store information of order in a seperated domain
func NewOrderList(price *big.Int, orderTree *OrderTree) *OrderList {
	item := &OrderListItem{
		// HeadOrder: nil,
		// TailOrder: nil,
		// set to default common.Hash
		HeadOrder: EmptyKey(),
		TailOrder: EmptyKey(),
		Length:    0,
		Volume:    Zero(),
		Price:     CloneBigInt(price),
	}

	return NewOrderListWithItem(item, orderTree)

}

func NewOrderListWithItem(item *OrderListItem, orderTree *OrderTree) *OrderList {
	key := orderTree.getKeyFromPrice(item.Price)

	orderList := &OrderList{
		Item:      item,
		Key:       key,
		orderTree: orderTree,
	}

	// priceKey will be slot of order tree + plus price key
	// we can use orderList slot as orderbook slot to store sequential of orders
	if orderTree.orderBook != nil {
		orderList.slot = orderTree.orderBook.slot
	} else {
		// orderList.slot = Zero()
		//
		orderList.slot = new(big.Int).SetBytes(crypto.Keccak256(key))
	}

	return orderList
}

// func (orderList *OrderList) HeadOrderKey(keys ...[]byte) []byte {
// 	if len(keys) == 1 {
// 		orderList.Item.HeadOrder = keys[0]
// 	}

// 	return orderList.Item.HeadOrder
// }

func (orderList *OrderList) GetOrder(key []byte) *Order {
	// re-use method from orderbook, because orderlist has the same slot as orderbook
	return orderList.orderTree.orderBook.GetOrder(key)

	// if orderList.isEmptyKey(key) {
	// 	return nil
	// }
	// // orderID := key
	// storedKey := orderList.GetOrderIDFromKey(key)
	// orderItem := &OrderItem{}
	// // var orderItem *OrderItem
	// val, err := orderList.orderTree.orderDB.Get(storedKey, orderItem)
	// if err != nil {
	// 	fmt.Printf("Key not found :%x, %v\n", storedKey, err)
	// 	return nil
	// }

	// // orderItem := &OrderItem{}
	// // rlp.DecodeBytes(bytes, orderItem)
	// // orderItem := val.(*OrderItem)
	// order := &Order{
	// 	Item: val.(*OrderItem),
	// 	Key:  key,
	// }
	// return order

	// order := orderList.orderTree.GetOrder(orderID)
	// if order != nil {
	// 	// return the original key
	// 	order.Key = key
	// }

	// return order
}

func (orderList *OrderList) isEmptyKey(key []byte) bool {
	return orderList.orderTree.PriceTree.IsEmptyKey(key)
}

func (orderList *OrderList) Head() *Order {
	return orderList.GetOrder(orderList.Item.HeadOrder)
}

func (orderList *OrderList) Tail() *Order {
	return orderList.GetOrder(orderList.Item.TailOrder)
}

// String : travel the list to print it in nice format
func (orderList *OrderList) String(startDepth int) string {

	if orderList == nil {
		return "<nil>"
	}

	var buffer bytes.Buffer
	tabs := strings.Repeat("\t", startDepth)
	// buffer.WriteString(fmt.Sprintf("{\n\t%sHeadKey: %x\n\t%sTailKey: %x\n\t%sLength: %d\n\t%sVolume: %v\n\t%sPrice: %v",
	// tabs, orderList.Item.HeadOrder, tabs, orderList.Item.TailOrder,
	buffer.WriteString(fmt.Sprintf("{\n\t%sLength: %d\n\t%sVolume: %v\n\t%sPrice: %v",
		tabs, orderList.Item.Length, tabs, orderList.Item.Volume, tabs, orderList.Item.Price))

	buffer.WriteString("\n\t")
	buffer.WriteString(tabs)
	buffer.WriteString("Head:")
	linkedList := orderList.Head()
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
	if depth == 0 {
		buffer.WriteString(" <nil>")
	}
	buffer.WriteString("\n\t")
	buffer.WriteString(tabs)
	buffer.WriteString("Tail:")
	linkedList = orderList.Tail()
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
	if depth == 0 {
		buffer.WriteString(" <nil>")
	}
	buffer.WriteString("\n")
	buffer.WriteString(tabs)
	buffer.WriteString("}")
	return buffer.String()
}

// Less : compare if this order list is less than compared object
func (orderList *OrderList) Less(than *OrderList) bool {
	// cast to OrderList pointer
	return IsStrictlySmallerThan(orderList.Item.Price, than.Item.Price)
}

func (orderList *OrderList) Save() error {
	// value, err := orderList.orderTree.orderDB.EncodeToBytes(orderList.Item)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// // we use orderlist db file seperated from order
	// // orderList.db.Put(orderList.Key, value)
	// if orderList.orderTree.PriceTree.Debug {
	// 	fmt.Printf("Save orderlist key %x, value :%x\n", orderList.Key, value)
	// }
	// // fmt.Println("AFTER UPDATE", orderList.String(0))
	// return orderList.orderTree.PriceTree.Put(orderList.Key, value)

	return orderList.orderTree.SaveOrderList(orderList)
}

// return the input orderID
func (orderList *OrderList) GetOrderIDFromList(key []byte) uint64 {
	orderSlot := new(big.Int).SetBytes(key)
	// fmt.Println("FAIL", key, orderList.slot)
	return Sub(orderSlot, orderList.slot).Uint64()
	// orderbook:[1,2,3,4]
	// return key
}

// GetOrderIDFromKey
// If we allow the same orderid belongs to many pricelist, we must use slot
// otherwise just use 1 db for storing all orders of all pricelists
// currently we use auto increase ment id so no need slot
func (orderList *OrderList) GetOrderIDFromKey(key []byte) []byte {
	orderSlot := new(big.Int).SetBytes(key)
	// fmt.Println("FAIL", key, orderList.slot)
	return common.BigToHash(Add(orderList.slot, orderSlot)).Bytes()
	// orderbook:[1,2,3,4]
	// return key
}

// GetOrderID return the real slot key of order in this linked list
func (orderList *OrderList) GetOrderID(order *Order) []byte {
	return orderList.GetOrderIDFromKey(order.Key)
}

// OrderExist search order in orderlist
func (orderList *OrderList) OrderExist(key []byte) bool {
	// orderKey := key
	orderKey := orderList.GetOrderIDFromKey(key)
	found, _ := orderList.orderTree.orderDB.Has(orderKey)
	return found
}

// // OrderExist search order in orderlist
// func (orderList *OrderList) OrderExist(order *Order) bool {
// 	key := orderList.GetOrderID(order)
// 	found, _ := orderList.orderTree.OrderDB.Has(key)
// 	return found
// }

func (orderList *OrderList) SaveOrder(order *Order) error {
	// value, err := rlp.EncodeToBytes(order.Item)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }

	// using other db to store Order object
	// key := common.BytesToHash(order.Key).Bytes()
	// key := order.Key
	key := orderList.GetOrderID(order)
	if orderList.orderTree.orderDB.Debug {
		fmt.Printf("Save order key : %x, value :%s\n", key, ToJSON(order.Item))
	}

	return orderList.orderTree.orderDB.Put(key, order.Item)

}

// AppendOrder : append order into the order list
func (orderList *OrderList) AppendOrder(order *Order) error {

	if orderList.Item.Length == 0 {
		order.Item.NextOrder = EmptyKey()
		order.Item.PrevOrder = EmptyKey()
	} else {
		order.Item.PrevOrder = orderList.Item.TailOrder
		order.Item.NextOrder = EmptyKey()
	}

	// save into database first
	err := orderList.SaveOrder(order)
	if err != nil {
		return err
	}

	if orderList.Item.Length == 0 {
		orderList.Item.HeadOrder = order.Key
		orderList.Item.TailOrder = order.Key
	} else {
		tailOrder := orderList.GetOrder(orderList.Item.TailOrder)
		if tailOrder != nil {
			tailOrder.Item.NextOrder = order.Key
			orderList.Item.TailOrder = order.Key
			orderList.SaveOrder(tailOrder)
		}
	}
	orderList.Item.Length++
	orderList.Item.Volume = Add(orderList.Item.Volume, order.Item.Quantity)
	// fmt.Println("orderlist", orderList.String(0))
	return orderList.Save()
}

func (orderList *OrderList) DeleteOrder(order *Order) error {
	key := orderList.GetOrderID(order)
	return orderList.orderTree.orderDB.Delete(key, false)
}

// RemoveOrder : remove order from the order list
func (orderList *OrderList) RemoveOrder(order *Order) error {

	// fmt.Println("OrderItem", ToJSON(orderList.Item))
	if orderList.Item.Length == 0 {
		// empty mean nothing to delete
		return nil
	}

	err := orderList.DeleteOrder(order)

	if err != nil {
		// stop other operations
		return err
	}

	nextOrder := orderList.GetOrder(order.Item.NextOrder)
	prevOrder := orderList.GetOrder(order.Item.PrevOrder)

	// // if there is no Order
	// if nextOrder == nil && prevOrder == nil {
	// 	return nil
	// }

	// fmt.Println("DELETE", nextOrder, prevOrder, order)

	orderList.Item.Volume = Sub(orderList.Item.Volume, order.Item.Quantity)
	orderList.Item.Length--

	if nextOrder != nil && prevOrder != nil {
		nextOrder.Item.PrevOrder = prevOrder.Key
		prevOrder.Item.NextOrder = nextOrder.Key

		orderList.SaveOrder(nextOrder)
		orderList.SaveOrder(prevOrder)
	} else if nextOrder != nil {
		// this might be wrong
		nextOrder.Item.PrevOrder = EmptyKey()
		orderList.Item.HeadOrder = nextOrder.Key

		orderList.SaveOrder(nextOrder)
	} else if prevOrder != nil {
		prevOrder.Item.NextOrder = EmptyKey()
		orderList.Item.TailOrder = prevOrder.Key

		orderList.SaveOrder(prevOrder)
	} else {
		// empty
		orderList.Item.HeadOrder = EmptyKey()
		orderList.Item.TailOrder = EmptyKey()
	}

	// fmt.Println("AFTER DELETE", orderList.String(0))

	return orderList.Save()
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
