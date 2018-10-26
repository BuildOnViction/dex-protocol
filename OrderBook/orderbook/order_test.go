package orderbook

import (
	"strconv"
	"testing"
)

func TestNewOrder(t *testing.T) {
	var orderList OrderList
	dummyOrder := make(map[string]string)
	dummyOrder["timestamp"] = strconv.Itoa(testTimestamp)
	dummyOrder["quantity"] = testQuanity.String()
	dummyOrder["price"] = testPrice.String()
	dummyOrder["order_id"] = strconv.Itoa(testOrderID)
	dummyOrder["trade_id"] = strconv.Itoa(testTradeID)

	order := NewOrder(dummyOrder, &orderList)

	t.Logf("Order : %s", order)

	if !(order.Timestamp == testTimestamp) {
		t.Errorf("Timesmape incorrect, got: %d, want: %d.", order.Timestamp, testTimestamp)
	}

	if !(order.Quantity.Equal(testQuanity)) {
		t.Errorf("quantity incorrect, got: %d, want: %d.", order.Quantity, testQuanity)
	}

	if !(order.Price.Equal(testPrice)) {
		t.Errorf("price incorrect, got: %d, want: %d.", order.Price, testPrice)
	}

	if !(order.OrderID == strconv.Itoa(testOrderID)) {
		t.Errorf("order id incorrect, got: %s, want: %d.", order.OrderID, testOrderID)
	}

	if !(order.TradeID == strconv.Itoa(testTradeID)) {
		t.Errorf("trade id incorrect, got: %s, want: %d.", order.TradeID, testTradeID)
	}
}

func TestOrder(t *testing.T) {
	orderList := NewOrderList(testPrice)

	dummyOrder := make(map[string]string)
	dummyOrder["timestamp"] = strconv.Itoa(testTimestamp)
	dummyOrder["quantity"] = testQuanity.String()
	dummyOrder["price"] = testPrice.String()
	dummyOrder["order_id"] = strconv.Itoa(testOrderID)
	dummyOrder["trade_id"] = strconv.Itoa(testTradeID)

	order := NewOrder(dummyOrder, orderList)
	orderList.AppendOrder(order)
	order.UpdateQuantity(testQuanity1, testTimestamp1)

	if !(order.Quantity.Equal(testQuanity1)) {
		t.Errorf("order id incorrect, got: %s, want: %d.", order.OrderID, testOrderID)
	}

	if !(order.Timestamp == testTimestamp1) {
		t.Errorf("trade id incorrect, got: %s, want: %d.", order.TradeID, testTradeID)
	}

	// log in json format
	for i := 0; i < 10; i++ {
		dummyOrder1 := make(map[string]string)
		dummyOrder1["timestamp"] = strconv.Itoa(testTimestamp1)
		dummyOrder1["quantity"] = testQuanity1.String()
		dummyOrder1["price"] = testPrice1.String()
		dummyOrder1["order_id"] = strconv.Itoa(testOrderID1)
		dummyOrder1["trade_id"] = strconv.Itoa(testTradeID1)

		order1 := NewOrder(dummyOrder1, orderList)
		orderList.AppendOrder(order1)
	}

	t.Logf("Order List : %s", orderList.String(0))
}
