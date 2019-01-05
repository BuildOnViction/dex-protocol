package orderbook

import (
	"strconv"
	"testing"

	"github.com/shopspring/decimal"
)

func TestNewOrderTree(t *testing.T) {
	orderTree := testOrderTree

	dummyOrder := make(map[string]string)
	dummyOrder["timestamp"] = strconv.FormatUint(testTimestamp, 10)
	dummyOrder["quantity"] = testQuanity.String()
	dummyOrder["price"] = testPrice.String()
	dummyOrder["order_id"] = strconv.Itoa(testOrderID)
	dummyOrder["trade_id"] = strconv.Itoa(testTradeID)

	dummyOrder1 := make(map[string]string)
	dummyOrder1["timestamp"] = strconv.FormatUint(testTimestamp1, 10)
	dummyOrder1["quantity"] = testQuanity1.String()
	dummyOrder1["price"] = testPrice1.String()
	dummyOrder1["order_id"] = strconv.Itoa(testOrderID1)
	dummyOrder1["trade_id"] = strconv.Itoa(testTradeID1)

	dummyOrder2 := make(map[string]string)
	dummyOrder2["timestamp"] = strconv.FormatUint(testTimestamp2, 10)
	dummyOrder2["quantity"] = testQuanity2.String()
	dummyOrder2["price"] = testPrice2.String()
	dummyOrder2["order_id"] = strconv.Itoa(testOrderID2)
	dummyOrder2["trade_id"] = strconv.Itoa(testTradeID2)

	dummyOrder3 := make(map[string]string)
	dummyOrder3["timestamp"] = strconv.FormatUint(testTimestamp3, 10)
	dummyOrder3["quantity"] = testQuanity3.String()
	dummyOrder3["price"] = testPrice3.String()
	dummyOrder3["order_id"] = strconv.Itoa(testOrderID3)
	dummyOrder3["trade_id"] = strconv.Itoa(testTradeID3)

	if !(orderTree.Item.Volume.Equal(decimal.Zero)) {
		t.Errorf("orderTree.Volume incorrect, got: %d, want: %d.", orderTree.Item.Volume, decimal.Zero)
	}

	// if !(orderTree.NotEmpty()) {
	// 	t.Errorf("orderTree.Length() incorrect, got: %d, want: %d.", orderTree.NotEmpty(), 0)
	// }

	orderTree.InsertOrder(dummyOrder)
	orderTree.InsertOrder(dummyOrder1)

	if !(orderTree.PriceExist(testPrice)) {
		t.Errorf("orderTree does not contain price %d.", testPrice)
	}

	if !(orderTree.PriceExist(testPrice1)) {
		t.Errorf("orderTree does not contain price %d.", testPrice1)
	}

	if !(orderTree.Item.NumOrders == 2) {
		t.Errorf("orderTree.NumOrders incorrect, got: %d, want: %d.", orderTree.Item.NumOrders, 2)
	}

	orderTree.RemoveOrderByID([]byte(dummyOrder1["order_id"]))
	orderTree.RemoveOrderByID([]byte(dummyOrder["order_id"]))

	orderTree.InsertOrder(dummyOrder)
	orderTree.InsertOrder(dummyOrder1)
	orderTree.InsertOrder(dummyOrder2)
	orderTree.InsertOrder(dummyOrder3)

	if !(orderTree.MaxPrice().Equal(testPrice3)) {
		t.Errorf("orderTree.MaxPrice incorrect, got: %d, want: %d.", orderTree.MaxPrice(), testPrice3)
	}

	if !(orderTree.MinPrice().Equal(testPrice)) {
		t.Errorf("orderTree.MinPrice incorrect, got: %d, want: %d.", orderTree.MinPrice(), testPrice)
	}

	orderTree.RemovePrice(testPrice)

	if orderTree.PriceExist(testPrice) {
		t.Errorf("orderTree.MinPrice incorrect, got: %d, want: %d.", orderTree.MinPrice(), testPrice)
	}

	t.Logf("OrderTree : %s", orderTree.String(0))

	// TODO Check PriceList as well and verify with the orders
}
