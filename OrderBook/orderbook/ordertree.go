package orderbook

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	// rbt "github.com/emirpasic/gods/trees/redblacktree"
)

type OrderTreeItem struct {
	Volume    *big.Int `json:"volume"`    // Contains total quantity from all Orders in tree
	NumOrders uint64   `json:"numOrders"` // Contains count of Orders in tree
	// Depth         uint64   `json:"depth"`         // Number of different prices in tree (http://en.wikipedia.org/wiki/Order_book_(trading)#Book_depth)
	PriceTreeKey  []byte `json:"priceTreeKey"`  // Root Key of price tree
	PriceTreeSize uint64 `json:"priceTreeSize"` // Number of nodes, currently it is Depth
}

// OrderTree : order tree structure for travelling
type OrderTree struct {
	PriceTree *RedBlackTreeExtended `json:"priceTree"`
	// PriceMap  map[string]*OrderList `json:"priceMap"`  // Dictionary containing price : OrderList object
	// OrderMap  map[string]*Order     `json:"orderMap"`  // Dictionary containing order_id : Order object
	orderBook *OrderBook
	orderDB   *BatchDatabase // this is for order
	slot      *big.Int
	Key       []byte
	Item      *OrderTreeItem

	// orderListCache *lru.Cache // Cache for the recent orderList
}

// NewOrderTree create new order tree
func NewOrderTree(orderDB *BatchDatabase, key []byte, orderBook *OrderBook) *OrderTree {
	// create priceTree from db for order list
	// orderListDBPath := path.Join(datadir, "pricetree")
	// orderDBPath := path.Join(datadir, "order")
	priceTree := NewRedBlackTreeExtended(orderDB)
	// priceTree.Debug = orderDB.Debug

	// itemCache, _ := lru.New(defaultCacheLimit)
	// orderDB, _ := ethdb.NewLDBDatabase(orderDBPath, 0, 0)

	item := &OrderTreeItem{
		Volume:    Zero(),
		NumOrders: 0,
		// Depth:     0,
		PriceTreeSize: 0,
	}

	slot := new(big.Int).SetBytes(key)

	// we will need a lru for cache hit, and internal cache for orderbook db to do the batch update
	orderTree := &OrderTree{
		orderDB:   orderDB,
		PriceTree: priceTree,
		Key:       key,
		slot:      slot,
		Item:      item,
		orderBook: orderBook,
		// orderListCache: itemCache,
	}

	// must restore from db first to make sure we get corrent information
	// orderTree.Restore()
	// then update PriceTree after restore the order tree

	// update price tree
	// orderTree.PriceTree = priceTree
	// orderTree.Key = GetKeyFromBig(orderTree.slot)
	return orderTree
}

// we use hash as offset to store order tree information
// var OrderTreeKey = crypto.Keccak256([]byte("ordertree"))

// func (orderTree *OrderTree) DB() *BatchDatabase {
// 	return orderTree.orderDB
// }

func (orderTree *OrderTree) Save() error {
	// commit tree changes
	// orderTree.PriceTree.Commit()

	// update tree meta information, make sure item existed instead of checking rootKey
	priceTreeRoot := orderTree.PriceTree.Root()
	if priceTreeRoot != nil {
		orderTree.Item.PriceTreeKey = priceTreeRoot.Key
		orderTree.Item.PriceTreeSize = orderTree.Depth()
	}

	// using rlp.EncodeToBytes as underlying encode method
	// fmt.Printf("ordertree bytes save : %v\n", orderTree.Key)
	return orderTree.orderDB.Put(orderTree.Key, orderTree.Item)
	// ordertreeBytes, err := rlp.EncodeToBytes(orderTree.Item)
	// ordertreeBytes, err := json.Marshal(orderTree.Item)
	// fmt.Printf("ordertree bytes : %s, %x\n", ToJSON(orderTree.Item), ordertreeBytes)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }

	// return orderTree.OrderDB.Put(OrderTreeKey, ordertreeBytes)
}

// save this tree information then do database commit
func (orderTree *OrderTree) Commit() error {
	err := orderTree.Save()
	if err == nil {
		err = orderTree.orderDB.Commit()
	}
	return err
}

func (orderTree *OrderTree) Restore() error {
	// val.(*OrderTreeItem)
	val, err := orderTree.orderDB.Get(orderTree.Key, orderTree.Item)

	// fmt.Printf("ordertree bytes get : %v\n", orderTree.Key)
	if err == nil {
		// return rlp.DecodeBytes(ordertreeBytes, orderTree.Item)
		// return json.Unmarshal(ordertreeBytes, orderTree.Item)
		orderTree.Item = val.(*OrderTreeItem)

		// update root key for pricetree
		orderTree.PriceTree.SetRootKey(orderTree.Item.PriceTreeKey, orderTree.Item.PriceTreeSize)
	}

	return err
}

func (orderTree *OrderTree) String(startDepth int) string {
	tabs := strings.Repeat("\t", startDepth)
	return fmt.Sprintf("{\n\t%sMinPriceList: %s\n\t%sMaxPriceList: %s\n\t%sVolume: %v\n\t%sNumOrders: %d\n\t%sDepth: %d\n%s}",
		tabs, orderTree.MinPriceList().String(startDepth+1), tabs, orderTree.MaxPriceList().String(startDepth+1), tabs,
		orderTree.Item.Volume, tabs, orderTree.Item.NumOrders, tabs, orderTree.Depth(), tabs)
}

func (orderTree *OrderTree) Length() uint64 {
	return orderTree.Item.NumOrders
}

// Check the order database is emtpy or not
func (orderTree *OrderTree) NotEmpty() bool {
	// return len(orderTree.OrderMap)
	// iter := orderTree.OrderDB.NewIterator()
	// return iter.First()
	return orderTree.Item.NumOrders > 0
}

func (orderTree *OrderTree) GetOrder(key []byte, price *big.Int) *Order {
	orderList := orderTree.PriceList(price)
	if orderList == nil {
		return nil
	}

	// we can use orderID incremental way, so we just need a big slot from price of order tree
	return orderList.GetOrder(key)
	// bytes, err := orderTree.OrderDB.Get(key)
	// if err != nil {
	// 	fmt.Printf("Key not found :%x, %v\n", key, err)
	// 	return nil
	// }
	// orderItem := &OrderItem{}
	// rlp.DecodeBytes(bytes, orderItem)
	// order := &Order{
	// 	Item: orderItem,
	// 	Key:  key,
	// }
	// return order
}

// // Order : get the order by orderID
// func (orderTree *OrderTree) Order(orderID string) *Order {
// 	return orderTree.OrderMap[orderID]
// }

func (orderTree *OrderTree) getSlotFromPrice(price *big.Int) *big.Int {
	// orderListKey, _ := price.GobEncode(
	return Add(orderTree.slot, price)
	// return crypto.Keccak256(orderTree.slot.Bytes(), common.BigToHash(price).Bytes())
	// orderListKey := orderTree.getKeyFromPrice(price)
	// return new(big.Int).SetBytes(orderListKey)
}

// next time this price will be big.Int
func (orderTree *OrderTree) getKeyFromPrice(price *big.Int) []byte {
	// orderListKey, _ := price.GobEncode()
	orderListKey := orderTree.getSlotFromPrice(price)
	return GetKeyFromBig(orderListKey)
	// price is like index of array, so it is faster to calculate with hash ordertree.price = [1,4,5]
	// so we use hash(key . subkey)
	// return crypto.Keccak256(orderTree.Key, GetKeyFromBig(price))
}

// PriceList : get the price list from the price map using price as key
func (orderTree *OrderTree) PriceList(price *big.Int) *OrderList {
	// this will be wrong, we must return existing orderList
	// orderList := NewOrderList(price, orderTree)

	// cache is seperated for each ordertree, so no need to add the slot
	// cacheKey := price.String()
	// // cache hit
	// if cached, ok := orderTree.orderListCache.Get(cacheKey); ok {
	// 	fmt.Println("Cache hit")
	// 	return cached.(*OrderList)
	// }

	key := orderTree.getKeyFromPrice(price)
	bytes, found := orderTree.PriceTree.Get(key)

	if found {
		// update Item
		// rlp.DecodeBytes(bytes, orderList.Item)
		// return orderList

		orderList := orderTree.decodeOrderList(bytes)

		// // update cache
		// orderTree.orderListCache.Add(cacheKey, orderList)

		return orderList
	}

	return nil
	// return orderTree.PriceMap[price.String()]
}

// CreatePrice : create new price list into PriceTree and PriceMap
func (orderTree *OrderTree) CreatePrice(price *big.Int) *OrderList {

	// orderTree.Item.Depth++
	newList := NewOrderList(price, orderTree)
	// put new price list into tree
	newList.Save()

	// should use batch to optimize the performance
	orderTree.Save()

	// // update cache
	// orderTree.orderListCache.Add(price.String(), newList)

	return newList
}

func (orderTree *OrderTree) SaveOrderList(orderList *OrderList) error {
	value, err := orderTree.orderDB.EncodeToBytes(orderList.Item)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// we use orderlist db file seperated from order
	// orderList.db.Put(orderList.Key, value)
	if orderTree.orderDB.Debug {
		fmt.Printf("Save orderlist key %x, value :%x\n", orderList.Key, value)
	}
	// fmt.Println("AFTER UPDATE", orderList.String(0))
	return orderTree.PriceTree.Put(orderList.Key, value)

}

func (orderTree *OrderTree) Depth() uint64 {
	return orderTree.PriceTree.Size()
}

// RemovePrice : delete a list by price
func (orderTree *OrderTree) RemovePrice(price *big.Int) {
	if orderTree.Depth() > 0 {
		// orderTree.Item.Depth--
		// using tree size
		orderListKey := orderTree.getKeyFromPrice(price)
		orderTree.PriceTree.Remove(orderListKey)

		// // also remove from cache to trigger cache miss
		// orderTree.orderListCache.Remove(price.String())

		// should use batch to optimize the performance
		orderTree.Save()
	}
}

// PriceExist : check price existed
func (orderTree *OrderTree) PriceExist(price *big.Int) bool {

	// cache hit
	// if orderTree.orderListCache.Contains(price.String()) {
	// 	return true
	// }

	orderListKey := orderTree.getKeyFromPrice(price)

	found, _ := orderTree.PriceTree.Has(orderListKey)
	// if found != true {
	// 	fmt.Println("FOUND", hex.EncodeToString(orderListKey), price.String())
	// }
	// fmt.Printf("Key :%x, %s, %x", orderListKey, price.String(), value)
	return found
}

// OrderExist : check order existed, only support for a specific price
func (orderTree *OrderTree) OrderExist(key []byte, price *big.Int) bool {
	orderList := orderTree.PriceList(price)
	if orderList == nil {
		return false
	}

	return orderList.OrderExist(key)
}

// InsertOrder : insert new order using quote data as map
func (orderTree *OrderTree) InsertOrder(quote map[string]string) error {

	// orderID := ToBigInt(quote["order_id"])
	// key := GetKeyFromBig(orderID)

	// if orderTree.OrderExist(key) {
	// 	// orderTree.RemoveOrderByID(key)
	// 	fmt.Println("Order already exsited, do nothing or should remove it?")
	// 	return
	// }

	price := ToBigInt(quote["price"])

	var orderList *OrderList

	if !orderTree.PriceExist(price) {
		// create and save
		fmt.Println("CREATE price list", price.String())
		orderList = orderTree.CreatePrice(price)
	} else {
		orderList = orderTree.PriceList(price)
	}

	// order will be insert if there is a follow orderList key
	if orderList != nil {

		order := NewOrder(quote, orderList.Key)

		if orderList.OrderExist(order.Key) {
			// orderTree.RemoveOrderByID(key)
			orderTree.RemoveOrder(order)
			// fmt.Println("Order already exsited, do nothing or should remove it?")
			// return nil
		}

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
		return orderTree.Save()
	}

	return nil
}

// UpdateOrder : update an order
func (orderTree *OrderTree) UpdateOrder(quote map[string]string) error {
	// order := orderTree.OrderMap[quote["order_id"]]

	price := ToBigInt(quote["price"])
	orderList := orderTree.PriceList(price)

	if orderList == nil {
		// create a price list for this price
		orderList = orderTree.CreatePrice(price)
	}

	orderID := ToBigInt(quote["order_id"])
	key := GetKeyFromBig(orderID)
	// order := orderTree.GetOrder(key)

	order := orderList.GetOrder(key)

	originalQuantity := CloneBigInt(order.Item.Quantity)

	if !IsEqual(price, order.Item.Price) {
		// Price changed. Remove order and update tree.
		// orderList := orderTree.PriceMap[order.Price.String()]
		orderList.RemoveOrder(order)
		if orderList.Item.Length == 0 {
			orderTree.RemovePrice(price)
		}
		orderTree.InsertOrder(quote)
		// orderList.Save()
	} else {
		quantity := ToBigInt(quote["quantity"])
		timestamp, _ := strconv.ParseUint(quote["timestamp"], 10, 64)
		order.UpdateQuantity(orderList, quantity, timestamp)
	}

	// fmt.Println("QUANTITY", order.Item.Quantity.String())

	orderTree.Item.Volume = Add(orderTree.Item.Volume, Sub(order.Item.Quantity, originalQuantity))

	// should use batch to optimize the performance
	return orderTree.Save()
}

func (orderTree *OrderTree) RemoveOrderFromOrderList(order *Order, orderList *OrderList) error {
	// next update orderList
	err := orderList.RemoveOrder(order)

	if err != nil {
		return err
	}

	// no items left than safety remove
	if orderList.Item.Length == 0 {
		orderTree.RemovePrice(order.Item.Price)
		fmt.Println("REMOVE price list", order.Item.Price.String())
	}

	// fmt.Println("QUANTITY", order.Item.Quantity.String())

	// update orderTree
	orderTree.Item.Volume = Sub(orderTree.Item.Volume, order.Item.Quantity)

	// delete(orderTree.OrderMap, orderID)
	orderTree.Item.NumOrders--

	// fmt.Println(orderTree.String(0))
	// fmt.Println("AFTER DELETE", orderList.String(0))

	// should use batch to optimize the performance
	return orderTree.Save()
}

func (orderTree *OrderTree) RemoveOrder(order *Order) (*OrderList, error) {
	// fmt.Printf("Node :%#v \n", order.Item)

	// then remove order from orderDB, 1 order can belong to muliple order price?
	// but we must not store order.meta in the same db, there must be one for order.meta, other for order.item
	// err := orderTree.OrderDB.Delete(order.Key)

	// if err != nil {
	// 	// stop other operations
	// 	return err
	// }
	var err error
	// get orderList by price, if there is orderlist, we will update it
	orderList := orderTree.PriceList(order.Item.Price)
	if orderList != nil {
		// next update orderList
		// err := orderList.RemoveOrder(order)

		// if err != nil {
		// 	return nil, err
		// }

		// // no items left than safety remove
		// if orderList.Item.Length == 0 {
		// 	orderTree.RemovePrice(order.Item.Price)
		// 	fmt.Println("REMOVE price list", order.Item.Price.String())
		// }

		// // update orderTree
		// orderTree.Item.Volume = Sub(orderTree.Item.Volume, order.Item.Quantity)

		// // delete(orderTree.OrderMap, orderID)
		// orderTree.Item.NumOrders--

		// // fmt.Println(orderTree.String(0))
		// // fmt.Println("AFTER DELETE", orderList.String(0))

		// // should use batch to optimize the performance
		// err = orderTree.Save()

		err = orderTree.RemoveOrderFromOrderList(order, orderList)

	}

	return orderList, err

}

// RemoveOrderByID : remove info using orderID
// func (orderTree *OrderTree) RemoveOrderByID(key []byte) error {

// 	// order := orderTree.OrderMap[orderID]
// 	order := orderTree.GetOrder(key)
// 	if order == nil {
// 		return nil
// 	}

// 	return orderTree.RemoveOrder(order)

// }

func (orderTree *OrderTree) getOrderListItem(bytes []byte) *OrderListItem {
	item := &OrderListItem{}
	// rlp.DecodeBytes(bytes, item)
	orderTree.orderDB.DecodeBytes(bytes, item)
	return item
}

func (orderTree *OrderTree) decodeOrderList(bytes []byte) *OrderList {
	item := orderTree.getOrderListItem(bytes)
	orderList := NewOrderListWithItem(item, orderTree)

	// // update cache
	// orderTree.orderListCache.Add(item.Price.String(), orderList)

	return orderList
	// return &OrderList{
	// 	Item:      item,
	// 	Key:       orderTree.getKeyFromPrice(item.Price),
	// 	orderTree: orderTree,
	// }
}

// MaxPrice : get the max price
func (orderTree *OrderTree) MaxPrice() *big.Int {
	if orderTree.Depth() > 0 {
		if bytes, found := orderTree.PriceTree.GetMax(); found {
			item := orderTree.getOrderListItem(bytes)
			if item != nil {
				return CloneBigInt(item.Price)
			}
		}
	}
	return Zero()
}

// MinPrice : get the min price
func (orderTree *OrderTree) MinPrice() *big.Int {
	if orderTree.Depth() > 0 {
		if bytes, found := orderTree.PriceTree.GetMin(); found {
			item := orderTree.getOrderListItem(bytes)
			if item != nil {
				return CloneBigInt(item.Price)
			}
		}
	}
	return Zero()
}

// MaxPriceList : get max price list
func (orderTree *OrderTree) MaxPriceList() *OrderList {
	if orderTree.Depth() > 0 {
		if bytes, found := orderTree.PriceTree.GetMax(); found {
			return orderTree.decodeOrderList(bytes)
		}
	}
	return nil

}

// MinPriceList : get min price list
func (orderTree *OrderTree) MinPriceList() *OrderList {
	if orderTree.Depth() > 0 {
		if bytes, found := orderTree.PriceTree.GetMin(); found {
			return orderTree.decodeOrderList(bytes)
		}
	}
	return nil
}
