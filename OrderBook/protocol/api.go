package protocol

import (
	"strconv"

	"github.com/ethereum/go-ethereum/swarm/storage/feed/lookup"

	demo "github.com/tomochain/orderbook/common"
)

// remember that API structs to be offered MUST be exported
type OrderbookAPI struct {
	V     int
	Model *OrderbookModel
}

// Version : return version
func (api *OrderbookAPI) Version() (int, error) {
	return api.V, nil
}

func NewOrderbookAPI(v int, orderbookModel *OrderbookModel) *OrderbookAPI {
	return &OrderbookAPI{
		V:     v,
		Model: orderbookModel,
	}
}

func (api *OrderbookAPI) GetBestAskList() []map[string]string {
	orderList := api.Model.Orderbook.Asks.MaxPriceList()
	if orderList == nil {
		return nil
	}
	// t.Logf("Best ask List : %s", orderList.String(0))
	cusor := orderList.HeadOrder
	// we have length
	var results []map[string]string
	for cusor != nil {
		record := make(map[string]string)
		record["timestamp"] = strconv.Itoa(cusor.Timestamp)
		record["price"] = cusor.Price.String()
		record["quantity"] = cusor.Quantity.String()
		record["order_id"] = cusor.OrderID
		record["trade_id"] = cusor.TradeID

		results = append(results, record)

		cusor = cusor.NextOrder
	}
	return results
}

func (api *OrderbookAPI) GetBestBidList() []map[string]string {
	orderList := api.Model.Orderbook.Bids.MinPriceList()
	// t.Logf("Best ask List : %s", orderList.String(0))
	if orderList == nil {
		return nil
	}
	cusor := orderList.TailOrder
	// we have length
	results := make([]map[string]string, orderList.Length)
	for cusor != nil {
		record := make(map[string]string)
		record["timestamp"] = strconv.Itoa(cusor.Timestamp)
		record["price"] = cusor.Price.String()
		record["quantity"] = cusor.Quantity.String()
		record["order_id"] = cusor.OrderID
		record["trade_id"] = cusor.TradeID

		results = append(results, record)

		cusor = cusor.PrevOrder
	}
	return results

}

func (api *OrderbookAPI) GetOrders(coin string, signerAddress string) []OrderbookMsg {
	messages, _ := api.Model.GetOrdersByAddress(coin, signerAddress)
	demo.LogInfo("Got data", "coin", coin, "address", signerAddress, "messages", messages)
	return messages
}

func (api *OrderbookAPI) UpdateOrders(coin string, signerAddress string, hexData, hexSignature string, time uint64, level uint8) error {

	epoch := lookup.Epoch{
		Time:  time,
		Level: level,
	}
	return api.Model.UpdateData(coin, signerAddress, epoch, hexData, hexSignature)
}
