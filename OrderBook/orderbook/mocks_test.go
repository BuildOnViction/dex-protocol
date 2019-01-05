package orderbook

import (
	"github.com/shopspring/decimal"
)

var testOrderTree = NewOrderTree("../datadir/testing")

var testTimestamp uint64 = 123452342343
var testQuanity, _ = decimal.NewFromString("0.1")
var testPrice, _ = decimal.NewFromString("0.1")
var testOrderID = 1
var testTradeID = 1

var testTimestamp1 uint64 = 123452342345
var testQuanity1, _ = decimal.NewFromString("0.2")
var testPrice1, _ = decimal.NewFromString("0.1")
var testOrderID1 = 2
var testTradeID1 = 2

var testTimestamp2 uint64 = 123452342340
var testQuanity2, _ = decimal.NewFromString("0.2")
var testPrice2, _ = decimal.NewFromString("0.3")
var testOrderID2 = 3
var testTradeID2 = 3

var testTimestamp3 uint64 = 1234523
var testQuanity3, _ = decimal.NewFromString("200.0")
var testPrice3, _ = decimal.NewFromString("1.3")
var testOrderID3 = 3
var testTradeID3 = 3
