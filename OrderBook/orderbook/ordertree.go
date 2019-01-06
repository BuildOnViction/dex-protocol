package orderbook

import (
	"fmt"
	"math/big"
	"path"
	"strconv"
	"strings"

	// rbt "github.com/emirpasic/gods/trees/redblacktree"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

type OrderTreeItem struct {
	Volume    *big.Int `json:"volume"`    // Contains total quantity from all Orders in tree
	NumOrders uint64   `json:"numOrders"` // Contains count of Orders in tree
	Depth     uint64   `json:"depth"`     // Number of different prices in tree (http://en.wikipedia.org/wiki/Order_book_(trading)#Book_depth)
}

// OrderTree : order tree structure for travelling
type OrderTree struct {
	PriceTree *RedBlackTreeExtended `json:"priceTree"`
	// PriceMap  map[string]*OrderList `json:"priceMap"`  // Dictionary containing price : OrderList object
	// OrderMap  map[string]*Order     `json:"orderMap"`  // Dictionary containing order_id : Order object

	OrderDB *ethdb.LDBDatabase // this is for order

	Item *OrderTreeItem
}

// NewOrderTree create new order tree
func NewOrderTree(datadir string) *OrderTree {
	// create priceTree from db for order list
	orderListDBPath := path.Join(datadir, "pricetree")
	orderDBPath := path.Join(datadir, "order")
	priceTree := NewRedBlackTreeExtended(orderListDBPath)
	orderDB, _ := ethdb.NewLDBDatabase(orderDBPath, 0, 0)

	item := &OrderTreeItem{
		Volume:    big.NewInt(0),
		NumOrders: 0,
		Depth:     0,
	}
	orderTree := &OrderTree{
		PriceTree: priceTree,
		OrderDB:   orderDB,
		Item:      item,
	}

	// must restore from db first to make sure we get corrent information
	orderTree.Restore()

	return orderTree
}

// we use hash as offset to store order tree information
var OrderTreeKey = crypto.Keccak256([]byte("ordertree"))

func (orderTree *OrderTree) Save() error {
	// commit tree changes
	orderTree.PriceTree.Commit()

	// update tree meta information
	ordertreeBytes, err := rlp.EncodeToBytes(orderTree.Item)
	// fmt.Printf("ordertree bytes : %s, %x\n", ToJSON(orderTree.Item), ordertreeBytes)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return orderTree.OrderDB.Put(OrderTreeKey, ordertreeBytes)
}

func (orderTree *OrderTree) Restore() error {
	ordertreeBytes, err := orderTree.OrderDB.Get(OrderTreeKey)
	// fmt.Printf("ordertree bytes : %x\n", ordertreeBytes)
	if err == nil {
		return rlp.DecodeBytes(ordertreeBytes, orderTree.Item)
	}
	return err
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
	return GetKeyFromBig(price)
}

// PriceList : get the price list from the price map using price as key
func (orderTree *OrderTree) PriceList(price *big.Int) *OrderList {
	orderListKey := orderTree.getKeyFromPrice(price)
	bytes, found := orderTree.PriceTree.Get(orderListKey)

	if found {
		item := &OrderListItem{}
		rlp.DecodeBytes(bytes, item)
		return &OrderList{
			Item:      item,
			Key:       orderListKey,
			orderTree: orderTree,
		}
	}

	return nil
	// return orderTree.PriceMap[price.String()]
}

// CreatePrice : create new price list into PriceTree and PriceMap
func (orderTree *OrderTree) CreatePrice(price *big.Int) *OrderList {
	orderTree.Item.Depth++
	newList := NewOrderList(price, orderTree)
	// put new price list into tree
	newList.Save()
	// should use batch to optimize the performance
	orderTree.Save()
	return newList
}

// RemovePrice : delete a list by price
func (orderTree *OrderTree) RemovePrice(price *big.Int) {
	if orderTree.Item.Depth > 0 {
		orderTree.Item.Depth--
		orderListKey := orderTree.getKeyFromPrice(price)
		orderTree.PriceTree.Remove(orderListKey)

		// should use batch to optimize the performance
		orderTree.Save()
	}
}

// PriceExist : check price existed
func (orderTree *OrderTree) PriceExist(price *big.Int) bool {
	orderListKey := orderTree.getKeyFromPrice(price)

	_, found := orderTree.PriceTree.Get(orderListKey)
	// fmt.Printf("Key :%x, %s, %x", orderListKey, price.String(), value)
	return found
}

// OrderExist : check order existed
func (orderTree *OrderTree) OrderExist(key []byte) bool {
	found, _ := orderTree.OrderDB.Has(key)
	return found
}

// InsertOrder : insert new order using quote data as map
func (orderTree *OrderTree) InsertOrder(quote map[string]string) {

	orderID := ToBigInt(quote["order_id"])
	key := GetKeyFromBig(orderID)

	if orderTree.OrderExist(key) {
		// orderTree.RemoveOrderByID(key)
		fmt.Println("Order already exsited, do nothing or should remove it?")
		return
	}

	price := ToBigInt(quote["price"])

	var orderList *OrderList

	if !orderTree.PriceExist(price) {
		// create and save
		orderList = orderTree.CreatePrice(price)
	} else {
		orderList = orderTree.PriceList(price)
	}

	if orderList != nil {
		order := NewOrder(quote, orderList.Key)
		orderList.AppendOrder(order)
		// orderTree.OrderMap[order.OrderID] = order
		orderList.Save()
		orderList.SaveOrder(order)
		orderTree.Item.Volume = Add(orderTree.Item.Volume, order.Item.Quantity)

		// increase num of orders, should be big.Int ?
		orderTree.Item.NumOrders++
		// fmt.Println("Num order", orderTree.Item.NumOrders)
		// update
		// should use batch to optimize the performance
		orderTree.Save()
	}

}

// UpdateOrder : update an order
func (orderTree *OrderTree) UpdateOrder(quote map[string]string) {
	// order := orderTree.OrderMap[quote["order_id"]]
	orderID := ToBigInt(quote["order_id"])
	key := GetKeyFromBig(orderID)

	order := orderTree.Order(key)
	originalQuantity := order.Item.Quantity
	price := ToBigInt(quote["price"])
	orderList := orderTree.PriceList(order.Item.Price)

	if orderList == nil {
		return
	}

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
		quantity := ToBigInt(quote["quantity"])
		timestamp, _ := strconv.ParseUint(quote["timestamp"], 10, 64)
		order.UpdateQuantity(orderList, quantity, timestamp)
		orderList.SaveOrder(order)
	}
	orderTree.Item.Volume = Add(orderTree.Item.Volume, Sub(order.Item.Quantity, originalQuantity))

	// should use batch to optimize the performance
	orderTree.Save()
}

// RemoveOrderByID : remove info using orderID
func (orderTree *OrderTree) RemoveOrderByID(key []byte) {

	// order := orderTree.OrderMap[orderID]
	order := orderTree.Order(key)
	if order == nil {
		return
	}

	// fmt.Printf("Node :%#v \n", order.Item)

	orderTree.Item.Volume = Sub(orderTree.Item.Volume, order.Item.Quantity)
	// get orderList by price
	orderList := orderTree.PriceList(order.Item.Price)
	if orderList != nil {
		orderList.RemoveOrder(order)
		// no items left than safety remove
		if orderList.Item.Length == 0 {
			orderTree.RemovePrice(order.Item.Price)
		}
		// delete(orderTree.OrderMap, orderID)
		orderTree.Item.NumOrders--
		// should use batch to optimize the performance
		orderTree.Save()
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
				return CloneBigInt(item.Price)
			}
		}
	}
	return big.NewInt(0)
}

// MinPrice : get the min price

func (orderTree *OrderTree) MinPrice() *big.Int {
	if orderTree.Item.Depth > 0 {
		if bytes, found := orderTree.PriceTree.GetMin(); found {
			item := orderTree.getOrderListItem(bytes)
			if item != nil {
				return CloneBigInt(item.Price)
			}
		}
	}
	return big.NewInt(0)
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
