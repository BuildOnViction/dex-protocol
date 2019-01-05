package orderbook

import (
	"bytes"
	"math/big"
	"strconv"
	"testing"

	"github.com/tomochain/backend-matching-engine/utils/math"
)

func TestNewOrder(t *testing.T) {

	dummyOrder := make(map[string]string)
	dummyOrder["timestamp"] = strconv.FormatUint(testTimestamp, 10)
	dummyOrder["quantity"] = testQuanity.String()
	dummyOrder["price"] = testPrice.String()
	dummyOrder["order_id"] = strconv.Itoa(testOrderID)
	dummyOrder["trade_id"] = strconv.Itoa(testTradeID)
	priceKey := testOrderTree.getKeyFromPrice(testPrice)
	order := NewOrder(dummyOrder, priceKey)

	t.Logf("Order : %s", order)

	if !(order.Item.Timestamp == testTimestamp) {
		t.Errorf("Timesmape incorrect, got: %d, want: %d.", order.Item.Timestamp, testTimestamp)
	}

	if order.Item.Quantity.Cmp(testQuanity) != 0 {
		t.Errorf("quantity incorrect, got: %d, want: %d.", order.Item.Quantity, testQuanity)
	}

	if order.Item.Price.Cmp(testPrice) != 0 {
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
	dummyOrder["timestamp"] = strconv.FormatUint(testTimestamp, 10)
	dummyOrder["quantity"] = testQuanity.String()
	dummyOrder["price"] = testPrice.String()
	dummyOrder["order_id"] = strconv.Itoa(testOrderID)
	dummyOrder["trade_id"] = strconv.Itoa(testTradeID)

	order := NewOrder(dummyOrder, orderList.Key)
	orderList.AppendOrder(order)
	order.UpdateQuantity(orderList, testQuanity1, testTimestamp1)

	if order.Item.Quantity.Cmp(testQuanity1) != 0 {
		t.Errorf("order id incorrect, got: %s, want: %d.", order.Key, testOrderID)
	}

	if !(order.Item.Timestamp == testTimestamp1) {
		t.Errorf("trade id incorrect, got: %s, want: %d.", order.Item.TradeID, testTradeID)
	}

	// log in json format
	var i int64 = 1
	for ; i < 10; i++ {
		increment := big.NewInt(i)
		dummyOrder1 := make(map[string]string)
		dummyOrder1["timestamp"] = strconv.FormatUint(testTimestamp1, 10)
		dummyOrder1["quantity"] = testQuanity1.String()
		dummyOrder1["price"] = math.Add(testPrice1, increment).String()
		dummyOrder1["order_id"] = dummyOrder1["price"]
		dummyOrder1["trade_id"] = strconv.Itoa(testTradeID1)

		order1 := NewOrder(dummyOrder1, orderList.Key)
		orderList.AppendOrder(order1)
	}

	t.Logf("Order List : %s", orderList.String(0))
}
