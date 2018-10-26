package orderbook

import (
	"fmt"
	"strconv"
	"strings"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/shopspring/decimal"
)

// Comparator : compare 2 interface
type Comparator func(a, b interface{}) int

func decimalComparator(a, b interface{}) int {
	aAsserted := a.(decimal.Decimal)
	bAsserted := b.(decimal.Decimal)
	switch {
	case aAsserted.GreaterThan(bAsserted):
		return 1
	case aAsserted.LessThan(bAsserted):
		return -1
	default:
		return 0
	}
}

// OrderTree : order tree structure for travelling
type OrderTree struct {
	PriceTree *RedBlackTreeExtended `json:"priceTree"`
	PriceMap  map[string]*OrderList `json:"priceMap"`  // Dictionary containing price : OrderList object
	OrderMap  map[string]*Order     `json:"orderMap"`  // Dictionary containing order_id : Order object
	Volume    decimal.Decimal       `json:"volume"`    // Contains total quantity from all Orders in tree
	NumOrders int                   `json:"numOrders"` // Contains count of Orders in tree
	Depth     int                   `json:"depth"`     // Number of different prices in tree (http://en.wikipedia.org/wiki/Order_book_(trading)#Book_depth)
}

// NewOrderTree create new order tree
func NewOrderTree() *OrderTree {
	priceTree := &RedBlackTreeExtended{rbt.NewWith(decimalComparator)}
	priceMap := make(map[string]*OrderList)
	orderMap := make(map[string]*Order)
	return &OrderTree{
		PriceTree: priceTree,
		PriceMap:  priceMap,
		OrderMap:  orderMap,
		Volume:    decimal.Zero,
		NumOrders: 0,
		Depth:     0,
	}
}

func (orderTree *OrderTree) String(startDepth int) string {
	tabs := strings.Repeat("\t", startDepth)
	return fmt.Sprintf("{\n\t%sMinPriceList: %s\n\t%sMaxPriceList: %s\n\t%sVolume: %v\n\t%sNumOrders: %d\n\t%sDepth: %d\n%s}",
		tabs, orderTree.MinPriceList().String(startDepth+1), tabs, orderTree.MaxPriceList().String(startDepth+1), tabs, orderTree.Volume, tabs, orderTree.NumOrders, tabs, orderTree.Depth, tabs)
}

// Length : get the length from the ordermap internally
func (orderTree *OrderTree) Length() int {
	return len(orderTree.OrderMap)
}

// Order : get the order by orderID
func (orderTree *OrderTree) Order(orderID string) *Order {
	return orderTree.OrderMap[orderID]
}

// PriceList : get the price list from the price map using price as key
func (orderTree *OrderTree) PriceList(price decimal.Decimal) *OrderList {
	return orderTree.PriceMap[price.String()]
}

// CreatePrice : create new price list into PriceTree and PriceMap
func (orderTree *OrderTree) CreatePrice(price decimal.Decimal) {
	orderTree.Depth++
	newList := NewOrderList(price)
	orderTree.PriceTree.Put(price, newList)
	orderTree.PriceMap[price.String()] = newList
}

// RemovePrice : delete a list by price
func (orderTree *OrderTree) RemovePrice(price decimal.Decimal) {
	orderTree.Depth--
	orderTree.PriceTree.Remove(price)
	delete(orderTree.PriceMap, price.String())
}

// PriceExist : check price existed
func (orderTree *OrderTree) PriceExist(price decimal.Decimal) bool {
	if _, ok := orderTree.PriceMap[price.String()]; ok {
		return true
	}
	return false
}

// OrderExist : check order existed
func (orderTree *OrderTree) OrderExist(orderID string) bool {
	if _, ok := orderTree.OrderMap[orderID]; ok {
		return true
	}
	return false
}

// InsertOrder : insert new order using quote data as map
func (orderTree *OrderTree) InsertOrder(quote map[string]string) {
	orderID := quote["order_id"]

	if orderTree.OrderExist(orderID) {
		orderTree.RemoveOrderByID(orderID)
	}
	orderTree.NumOrders++

	price, _ := decimal.NewFromString(quote["price"])

	if !orderTree.PriceExist(price) {
		orderTree.CreatePrice(price)
	}

	order := NewOrder(quote, orderTree.PriceMap[price.String()])
	orderTree.PriceMap[price.String()].AppendOrder(order)
	orderTree.OrderMap[order.OrderID] = order
	orderTree.Volume = orderTree.Volume.Add(order.Quantity)
}

// UpdateOrder : update an order
func (orderTree *OrderTree) UpdateOrder(quote map[string]string) {
	order := orderTree.OrderMap[quote["order_id"]]
	originalQuantity := order.Quantity
	price, _ := decimal.NewFromString(quote["price"])

	if !price.Equal(order.Price) {
		// Price changed. Remove order and update tree.
		orderList := orderTree.PriceMap[order.Price.String()]
		orderList.RemoveOrder(order)
		if orderList.Length == 0 {
			orderTree.RemovePrice(price)
		}
		orderTree.InsertOrder(quote)
	} else {
		quantity, _ := decimal.NewFromString(quote["quantity"])
		timestamp, _ := strconv.Atoi(quote["timestamp"])
		order.UpdateQuantity(quantity, timestamp)
	}
	orderTree.Volume = orderTree.Volume.Add(order.Quantity.Sub(originalQuantity))
}

// RemoveOrderByID : remove info using orderID
func (orderTree *OrderTree) RemoveOrderByID(orderID string) {
	orderTree.NumOrders--
	order := orderTree.OrderMap[orderID]
	orderTree.Volume = orderTree.Volume.Sub(order.Quantity)
	order.OrderList.RemoveOrder(order)
	// no items left than safety remove
	if order.OrderList.Length == 0 {
		orderTree.RemovePrice(order.Price)
	}
	delete(orderTree.OrderMap, orderID)
}

// MaxPrice : get the max price
func (orderTree *OrderTree) MaxPrice() decimal.Decimal {
	if orderTree.Depth > 0 {
		if value, found := orderTree.PriceTree.GetMax(); found {
			return value.(*OrderList).Price
		}
	}
	return decimal.Zero
}

// MinPrice : get the min price
func (orderTree *OrderTree) MinPrice() decimal.Decimal {
	if orderTree.Depth > 0 {
		if value, found := orderTree.PriceTree.GetMin(); found {
			return value.(*OrderList).Price
		}
	}
	return decimal.Zero
}

// MaxPriceList : get max price list
func (orderTree *OrderTree) MaxPriceList() *OrderList {
	if orderTree.Depth > 0 {
		price := orderTree.MaxPrice()
		return orderTree.PriceMap[price.String()]
	}
	return nil

}

// MinPriceList : get min price list
func (orderTree *OrderTree) MinPriceList() *OrderList {
	if orderTree.Depth > 0 {
		price := orderTree.MinPrice()
		return orderTree.PriceMap[price.String()]
	}
	return nil
}
