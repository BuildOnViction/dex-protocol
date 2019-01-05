package orderbook

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/shopspring/decimal"
)

func TestNewOrderList(t *testing.T) {
	orderList := NewOrderList(testPrice, testOrderTree)

	if !(orderList.Item.Length == 0) {
		t.Errorf("Orderlist length incorrect, got: %d, want: %d.", orderList.Item.Length, 0)
	}

	if !(orderList.Item.Price.Equal(testPrice)) {
		t.Errorf("Orderlist price incorrect, got: %d, want: %d.", orderList.Item.Price, testPrice)
	}

	if !(orderList.Item.Volume.Equal(decimal.Zero)) {
		t.Errorf("Orderlist volume incorrect, got: %d, want: %d.", orderList.Item.Volume, 0)
	}
}

func TestOrderList(t *testing.T) {
	orderList := NewOrderList(testPrice, testOrderTree)

	dummyOrder := make(map[string]string)
	dummyOrder["timestamp"] = strconv.FormatUint(testTimestamp, 10)
	dummyOrder["quantity"] = testQuanity.String()
	dummyOrder["price"] = testPrice.String()
	dummyOrder["order_id"] = strconv.Itoa(testOrderID)
	dummyOrder["trade_id"] = strconv.Itoa(testTradeID)

	order := NewOrder(dummyOrder, orderList.Key)
	orderList.AppendOrder(order)

	if !(orderList.Item.Length == 1) {
		t.Errorf("Orderlist Length incorrect, got: %d, want: %d.", orderList.Item.Length, 1)
	}

	if !(orderList.Item.Price.Equal(testPrice)) {
		t.Errorf("Orderlist price incorrect, got: %d, want: %d.", orderList.Item.Price, testPrice)
	}

	if !(orderList.Item.Volume.Equal(order.Item.Quantity)) {
		t.Errorf("Orderlist volume incorrect, got: %d, want: %d.", orderList.Item.Volume, order.Item.Quantity)
	}

	dummyOrder1 := make(map[string]string)
	dummyOrder1["timestamp"] = strconv.FormatUint(testTimestamp1, 10)
	dummyOrder1["quantity"] = testQuanity1.String()
	dummyOrder1["price"] = testPrice1.String()
	dummyOrder1["order_id"] = strconv.Itoa(testOrderID1)
	dummyOrder1["trade_id"] = strconv.Itoa(testTradeID1)

	order1 := NewOrder(dummyOrder1, orderList.Key)
	orderList.AppendOrder(order1)

	if !(orderList.Item.Length == 2) {
		t.Errorf("Orderlist Length incorrect, got: %d, want: %d.", orderList.Item.Length, 2)
	}

	orderListQuantity := order.Item.Quantity.Add(order1.Item.Quantity)
	if !(orderList.Item.Volume.Equal(orderListQuantity)) {
		t.Errorf("Orderlist Length incorrect, got: %d, want: %d.", orderList.Item.Volume, orderListQuantity)
	}

	headOrder := orderList.GetOrder(orderList.Item.HeadOrder)
	if !bytes.Equal(headOrder.Key, []byte("1")) {
		t.Errorf("headorder id incorrect, got: %x, want: %d.", headOrder.Key, 1)
	}

	nextOrder := orderList.GetOrder(headOrder.Item.NextOrder)

	if !bytes.Equal(nextOrder.Key, []byte("2")) {
		t.Errorf("Next headorder id incorrect, got: %x, want: %d.", nextOrder.Key, 2)
	}

	t.Logf("Order List : %s", orderList.String(0))
}
