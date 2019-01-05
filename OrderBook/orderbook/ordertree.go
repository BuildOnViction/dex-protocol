package orderbook

import (
	"fmt"
	"math/big"
	"path"
	"strconv"
	"strings"

	// rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/tomochain/backend-matching-engine/utils/math"
	rbt "github.com/tomochain/orderbook/redblacktree"
)

// // Comparator : compare 2 interface
// type Comparator func(a, b interface{}) int

// func decimalComparator(a, b interface{}) int {
// 	aAsserted := a.(decimal.Decimal)
// 	bAsserted := b.(decimal.Decimal)
// 	switch {
// 	case aAsserted.GreaterThan(bAsserted):
// 		return 1
// 	case aAsserted.LessThan(bAsserted):
// 		return -1
// 	default:
// 		return 0
// 	}
// }

type OrderTreeItem struct {
	Volume    *big.Int `json:"volume"`    // Contains total quantity from all Orders in tree
	NumOrders uint64   `json:"numOrders"` // Contains count of Orders in tree
	Depth     uint64   `json:"depth"`     // Number of different prices in tree (http://en.wikipedia.org/wiki/Order_book_(trading)#Book_depth)
}

// OrderTree : order tree structure for travelling
type OrderTree struct {
	PriceTree *rbt.RedBlackTreeExtended `json:"priceTree"`
	// PriceMap  map[string]*OrderList `json:"priceMap"`  // Dictionary containing price : OrderList object
	// OrderMap  map[string]*Order     `json:"orderMap"`  // Dictionary containing order_id : Order object

	OrderDB *ethdb.LDBDatabase // this is for order

	Item *OrderTreeItem
}

// NewOrderTree create new order tree
func NewOrderTree(datadir string) *OrderTree {
	// create priceTree from db for order list
	// priceTree := &RedBlackTreeExtended{rbt.NewWith(decimalComparator)}
	// priceMap := make(map[string]*OrderList)
	// orderMap := make(map[string]*Order)

	orderListDBPath := path.Join(datadir, "pricetree")
	orderDBPath := path.Join(datadir, "order")
	priceTree := rbt.NewRedBlackTreeExtended(orderListDBPath)
	orderDB, _ := ethdb.NewLDBDatabase(orderDBPath, 0, 0)

	item := &OrderTreeItem{
		Volume:    Zero,
		NumOrders: 0,
		Depth:     0,
	}
	return &OrderTree{
		PriceTree: priceTree,
		OrderDB:   orderDB,
		Item:      item,
	}
}

func (orderTree *OrderTree) String(startDepth int) string {
	tabs := strings.Repeat("\t", startDepth)
	return fmt.Sprintf("{\n\t%sMinPriceList: %s\n\t%sMaxPriceList: %s\n\t%sVolume: %v\n\t%sNumOrders: %d\n\t%sDepth: %d\n%s}",
		tabs, orderTree.MinPriceList().String(startDepth+1), tabs, orderTree.MaxPriceList().String(startDepth+1), tabs,
		orderTree.Item.Volume, tabs, orderTree.Item.NumOrders, tabs, orderTree.Item.Depth, tabs)
}

// Check the order database is emtpy or not
func (orderTree *OrderTree) NotEmpty() bool {
	// return len(orderTree.OrderMap)
	iter := orderTree.OrderDB.NewIterator()
	return iter.First()
}

func (orderTree *OrderTree) Order(key []byte) *Order {
	bytes, err := orderTree.OrderDB.Get(key)
	if err != nil {
		fmt.Printf("Key not found :%x, %s\n", key, key)
		return nil
	}
	orderItem := &OrderItem{}
	rlp.DecodeBytes(bytes, orderItem)
	order := &Order{
		Item: orderItem,
		Key:  key,
	}
	return order
}

// // Order : get the order by orderID
// func (orderTree *OrderTree) Order(orderID string) *Order {
// 	return orderTree.OrderMap[orderID]
// }

// next time this price will be big.Int
func (orderTree *OrderTree) getKeyFromPrice(price *big.Int) []byte {
	// orderListKey, _ := price.GobEncode()
	// return orderListKey
	// bigPrice := new(big.Int)
	// bigPrice.SetString(price.String(), 10)
	return common.BigToHash(price).Bytes()
}

// PriceList : get the price list from the price map using price as key
func (orderTree *OrderTree) PriceList(price *big.Int) *OrderList {
	orderListKey := orderTree.getKeyFromPrice(price)
	bytes, found := orderTree.PriceTree.Get(orderListKey)
	item := &OrderListItem{}

	if found {
		rlp.DecodeBytes(bytes, item)
	}

	return &OrderList{
		Item:      item,
		Key:       orderListKey,
		orderTree: orderTree,
	}
	// return orderTree.PriceMap[price.String()]
}

// CreatePrice : create new price list into PriceTree and PriceMap
func (orderTree *OrderTree) CreatePrice(price *big.Int) {
	orderTree.Item.Depth++
	newList := NewOrderList(price, orderTree)
	// orderTree.PriceTree.Put(price, newList)
	newList.Save()
	// orderTree.PriceMap[price.String()] = newList
}

// RemovePrice : delete a list by price
func (orderTree *OrderTree) RemovePrice(price *big.Int) {
	orderTree.Item.Depth--
	orderListKey := orderTree.getKeyFromPrice(price)
	orderTree.PriceTree.Remove(orderListKey)
}

// PriceExist : check price existed
func (orderTree *OrderTree) PriceExist(price *big.Int) bool {
	orderListKey := orderTree.getKeyFromPrice(price)
	// fmt.Printf("Key :%x, %s", orderListKey, price.String())
	_, found := orderTree.PriceTree.Get(orderListKey)
	return found
}

// OrderExist : check order existed
func (orderTree *OrderTree) OrderExist(key []byte) bool {
	found, _ := orderTree.OrderDB.Has(key)
	return found
}

// InsertOrder : insert new order using quote data as map
func (orderTree *OrderTree) InsertOrder(quote map[string]string) {
	key := []byte(quote["order_id"])

	if orderTree.OrderExist(key) {
		orderTree.RemoveOrderByID(key)
	}
	orderTree.Item.NumOrders++

	price, _ := new(big.Int).SetString(quote["price"], 10)

	if !orderTree.PriceExist(price) {
		orderTree.CreatePrice(price)
	}

	orderList := orderTree.PriceList(price)
	if orderList != nil {
		order := NewOrder(quote, orderList.Key)
		orderList.AppendOrder(order)
		// orderTree.OrderMap[order.OrderID] = order
		orderList.Save()
		orderList.SaveOrder(order)
		orderTree.Item.Volume = math.Add(orderTree.Item.Volume, order.Item.Quantity)
	}

}

// UpdateOrder : update an order
func (orderTree *OrderTree) UpdateOrder(quote map[string]string) {
	// order := orderTree.OrderMap[quote["order_id"]]
	key := []byte(quote["order_id"])
	order := orderTree.Order(key)
	originalQuantity := order.Item.Quantity
	price, _ := new(big.Int).SetString(quote["price"], 10)
	orderList := orderTree.PriceList(order.Item.Price)

	if price.Cmp(order.Item.Price) != 0 {
		// Price changed. Remove order and update tree.
		// orderList := orderTree.PriceMap[order.Price.String()]
		orderList.RemoveOrder(order)
		if orderList.Item.Length == 0 {
			orderTree.RemovePrice(price)
		}
		orderTree.InsertOrder(quote)
		orderList.Save()
	} else {
		quantity, _ := new(big.Int).SetString(quote["quantity"], 10)
		timestamp, _ := strconv.ParseUint(quote["timestamp"], 10, 64)
		order.UpdateQuantity(orderList, quantity, timestamp)
		orderList.SaveOrder(order)
	}
	orderTree.Item.Volume = math.Add(orderTree.Item.Volume, math.Sub(order.Item.Quantity, originalQuantity))
}

// RemoveOrderByID : remove info using orderID
func (orderTree *OrderTree) RemoveOrderByID(key []byte) {
	orderTree.Item.NumOrders--
	// order := orderTree.OrderMap[orderID]
	order := orderTree.Order(key)
	if order == nil {
		return
	}

	fmt.Printf("Node :%#v \n", order.Item)

	orderTree.Item.Volume = math.Sub(orderTree.Item.Volume, order.Item.Quantity)

	orderList := orderTree.PriceList(order.Item.Price)
	if orderList != nil {
		orderList.RemoveOrder(order)
		// no items left than safety remove
		if orderList.Item.Length == 0 {
			orderTree.RemovePrice(order.Item.Price)
		}
		// delete(orderTree.OrderMap, orderID)
	}
}

func (orderTree *OrderTree) getOrderListItem(bytes []byte) *OrderListItem {
	item := &OrderListItem{}
	rlp.DecodeBytes(bytes, item)
	return item
}

func (orderTree *OrderTree) DecodeOrderList(bytes []byte) *OrderList {
	item := orderTree.getOrderListItem(bytes)
	return &OrderList{
		Item:      item,
		Key:       orderTree.getKeyFromPrice(item.Price),
		orderTree: orderTree,
	}
}

// MaxPrice : get the max price
func (orderTree *OrderTree) MaxPrice() *big.Int {
	if orderTree.Item.Depth > 0 {
		if bytes, found := orderTree.PriceTree.GetMax(); found {
			item := orderTree.getOrderListItem(bytes)
			if item != nil {
				return item.Price
			}
		}
	}
	return Zero
}

// MinPrice : get the min price

func (orderTree *OrderTree) MinPrice() *big.Int {
	if orderTree.Item.Depth > 0 {
		if bytes, found := orderTree.PriceTree.GetMin(); found {
			item := orderTree.getOrderListItem(bytes)
			if item != nil {
				return item.Price
			}
		}
	}
	return Zero
}

// MaxPriceList : get max price list
func (orderTree *OrderTree) MaxPriceList() *OrderList {
	if orderTree.Item.Depth > 0 {
		if bytes, found := orderTree.PriceTree.GetMax(); found {
			return orderTree.DecodeOrderList(bytes)
		}
	}
	return nil

}

// MinPriceList : get min price list
func (orderTree *OrderTree) MinPriceList() *OrderList {
	if orderTree.Item.Depth > 0 {
		if bytes, found := orderTree.PriceTree.GetMin(); found {
			return orderTree.DecodeOrderList(bytes)
		}
	}
	return nil
}
