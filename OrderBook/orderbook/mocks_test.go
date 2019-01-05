package orderbook

import (
	"math/big"
)

var testOrderTree = NewOrderTree("../datadir/testing")

var testTimestamp uint64 = 123452342343
var testQuanity, _ = new(big.Int).SetString("1000", 10)
var testPrice, _ = new(big.Int).SetString("1000", 10)
var testOrderID = 1
var testTradeID = 1

var testTimestamp1 uint64 = 123452342345
var testQuanity1, _ = new(big.Int).SetString("2000", 10)
var testPrice1, _ = new(big.Int).SetString("1200", 10)
var testOrderID1 = 2
var testTradeID1 = 2

var testTimestamp2 uint64 = 123452342340
var testQuanity2, _ = new(big.Int).SetString("2000", 10)
var testPrice2, _ = new(big.Int).SetString("3000", 10)
var testOrderID2 = 3
var testTradeID2 = 3

var testTimestamp3 uint64 = 1234523
var testQuanity3, _ = new(big.Int).SetString("200000", 10)
var testPrice3, _ = new(big.Int).SetString("13000", 10)
var testOrderID3 = 3
var testTradeID3 = 3
