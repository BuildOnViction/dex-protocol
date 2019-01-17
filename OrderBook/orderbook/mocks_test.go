package orderbook

var datadir = "../datadir/testing"

// var datadir = "../../.data_30100/orderbook/"
// orderbook for this pair
var pairName = "TOMO/WETH"

// override Encode and Decode for better performance
var testDB = NewBatchDatabaseWithEncode(datadir, 0, 0, EncodeBytesItem, DecodeBytesItem)
var testOrderBook = NewOrderBook(pairName, testDB)

// order tree without orderbook
var testOrderTree = NewOrderTree(testDB, []byte("ordertree"), testOrderBook)

var testTimestamp uint64 = 123452342343
var testQuanity = ToBigInt("1000")
var testPrice = ToBigInt("1000")
var testOrderID = 1
var testTradeID = 1

var testTimestamp1 uint64 = 123452342345
var testQuanity1 = ToBigInt("2000")
var testPrice1 = ToBigInt("1200")
var testOrderID1 = 2
var testTradeID1 = 2

var testTimestamp2 uint64 = 123452342340
var testQuanity2 = ToBigInt("2000")
var testPrice2 = ToBigInt("3000")
var testOrderID2 = 3
var testTradeID2 = 3

var testTimestamp3 uint64 = 123452342347
var testQuanity3 = ToBigInt("200000")
var testPrice3 = ToBigInt("13000")
var testOrderID3 = 4
var testTradeID3 = 4
