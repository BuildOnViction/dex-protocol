package orderbook

import (
	"bytes"
	"strconv"
	"testing"
)

func TestNewOrder(t *testing.T) {

	dummyOrder := make(map[string]string)
	dummyOrder["timestamp"] = strconv.Itoa(testTimestamp)
	dummyOrder["quantity"] = testQuanity.String()
	dummyOrder["price"] = testPrice.String()
	dummyOrder["order_id"] = strconv.Itoa(testOrderID)
	dummyOrder["trade_id"] = strconv.Itoa(testTradeID)
	priceKey, _ := testPrice.GobEncode()
	order := NewOrder(dummyOrder, priceKey)

	t.Logf("Order : %s", order)

	if !(order.Item.Timestamp == testTimestamp) {
		t.Errorf("Timesmape incorrect, got: %d, want: %d.", order.Item.Timestamp, testTimestamp)
	}

	if !(order.Item.Quantity.Equal(testQuanity)) {
		t.Errorf("quantity incorrect, got: %d, want: %d.", order.Item.Quantity, testQuanity)
	}

	if !(order.Item.Price.Equal(testPrice)) {
		t.Errorf("price incorrect, got: %d, want: %d.", order.Item.Price, testPrice)
	}

	if !bytes.Equal(order.Key, []byte(dummyOrder["order_id"])) {
		t.Errorf("order id incorrect, got: %x, want: %d.", order.Key, testOrderID)
	}

	if !(order.Item.TradeID == strconv.Itoa(testTradeID)) {
		t.Errorf("trade id incorrect, got: %s, want: %d.", order.Item.TradeID, testTradeID)
	}
}

func TestOrder(t *testing.T) {
	orderList := NewOrderList(testPrice, testOrderTree)

	dummyOrder := make(map[string]string)
	dummyOrder["timestamp"] = strconv.Itoa(testTimestamp)
	dummyOrder["quantity"] = testQuanity.String()
	dummyOrder["price"] = testPrice.String()
	dummyOrder["order_id"] = strconv.Itoa(testOrderID)
	dummyOrder["trade_id"] = strconv.Itoa(testTradeID)

	order := NewOrder(dummyOrder, orderList.Key)
	orderList.AppendOrder(order)
	order.UpdateQuantity(orderList, testQuanity1, testTimestamp1)

	if !(order.Item.Quantity.Equal(testQuanity1)) {
		t.Errorf("order id incorrect, got: %s, want: %d.", order.Key, testOrderID)
	}

	if !(order.Item.Timestamp == testTimestamp1) {
		t.Errorf("trade id incorrect, got: %s, want: %d.", order.Item.TradeID, testTradeID)
	}

	// log in json format
	for i := 0; i < 10; i++ {
		dummyOrder1 := make(map[string]string)
		dummyOrder1["timestamp"] = strconv.Itoa(testTimestamp1)
		dummyOrder1["quantity"] = testQuanity1.String()
		dummyOrder1["price"] = testPrice1.String()
		dummyOrder1["order_id"] = strconv.Itoa(testOrderID1)
		dummyOrder1["trade_id"] = strconv.Itoa(testTradeID1)

		order1 := NewOrder(dummyOrder1, orderList.Key)
		orderList.AppendOrder(order1)
	}

	t.Logf("Order List : %s", orderList.String(0))
}
