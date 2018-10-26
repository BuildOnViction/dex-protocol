package orderbook

import (
	"strconv"
	"testing"

	"github.com/shopspring/decimal"
)

func TestNewOrderList(t *testing.T) {
	orderList := NewOrderList(testPrice)

	if !(orderList.Length == 0) {
		t.Errorf("Orderlist Length incorrect, got: %d, want: %d.", orderList.Length, 0)
	}

	if !(orderList.Price.Equal(testPrice)) {
		t.Errorf("Orderlist price incorrect, got: %d, want: %d.", orderList.Length, 0)
	}

	if !(orderList.Volume.Equal(decimal.Zero)) {
		t.Errorf("Orderlist Length incorrect, got: %d, want: %d.", orderList.Length, 0)
	}
}

func TestOrderList(t *testing.T) {
	orderList := NewOrderList(testPrice)

	dummyOrder := make(map[string]string)
	dummyOrder["timestamp"] = strconv.Itoa(testTimestamp)
	dummyOrder["quantity"] = testQuanity.String()
	dummyOrder["price"] = testPrice.String()
	dummyOrder["order_id"] = strconv.Itoa(testOrderID)
	dummyOrder["trade_id"] = strconv.Itoa(testTradeID)

	order := NewOrder(dummyOrder, orderList)
	orderList.AppendOrder(order)

	if !(orderList.Length == 1) {
		t.Errorf("Orderlist Length incorrect, got: %d, want: %d.", orderList.Length, 1)
	}

	if !(orderList.Price.Equal(testPrice)) {
		t.Errorf("Orderlist price incorrect, got: %d, want: %d.", orderList.Price, testPrice)
	}

	if !(orderList.Volume.Equal(order.Quantity)) {
		t.Errorf("Orderlist volume incorrect, got: %d, want: %d.", orderList.Volume, order.Quantity)
	}

	dummyOrder1 := make(map[string]string)
	dummyOrder1["timestamp"] = strconv.Itoa(testTimestamp1)
	dummyOrder1["quantity"] = testQuanity1.String()
	dummyOrder1["price"] = testPrice1.String()
	dummyOrder1["order_id"] = strconv.Itoa(testOrderID1)
	dummyOrder1["trade_id"] = strconv.Itoa(testTradeID1)

	order1 := NewOrder(dummyOrder1, orderList)
	orderList.AppendOrder(order1)

	if !(orderList.Length == 2) {
		t.Errorf("Orderlist Length incorrect, got: %d, want: %d.", orderList.Length, 2)
	}

	orderListQuantity := order.Quantity.Add(order1.Quantity)
	if !(orderList.Volume.Equal(orderListQuantity)) {
		t.Errorf("Orderlist Length incorrect, got: %d, want: %d.", orderList.Volume, orderListQuantity)
	}

	headOrder := orderList.HeadOrder
	if !(headOrder.OrderID == "1") {
		t.Errorf("headorder id incorrect, got: %s, want: %d.", headOrder.OrderID, 0)
	}

	nextOrder := headOrder.NextOrder

	if !(nextOrder.OrderID == "2") {
		t.Errorf("Next headorder id incorrect, got: %s, want: %d.", nextOrder.OrderID, 2)
	}

	t.Logf("Order List : %s", orderList.String(0))
}
