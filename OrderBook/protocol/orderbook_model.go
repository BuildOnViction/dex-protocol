package protocol

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/tomochain/orderbook/orderbook"
)

// const (
// 	Day       = 60 * 60 * 24
// 	Year      = Day * 365
// 	Month     = Day * 30
// 	TopicName = "Token"
// )

// OrderBookModel : singleton orderbook for testing
type OrderbookModel struct {
	Orderbooks map[string]*orderbook.OrderBook
	db         *orderbook.BatchDatabase
	// pair and max volume ...
	allowedPairs map[string]*big.Int
	// BzzClient *client.Client
	// Signer    *feed.GenericSigner
}

func NewModel(datadir string, allowedPairs map[string]*big.Int) *OrderbookModel {
	// demo.LogDebug("Creating model", "signerAddress", signer.Address().Hex())
	batchDB := orderbook.NewBatchDatabaseWithEncode(datadir, 0, 0,
		orderbook.EncodeBytesItem, orderbook.DecodeBytesItem)

	fixAllowedPairs := make(map[string]*big.Int)
	for key, value := range allowedPairs {
		fixAllowedPairs[strings.ToLower(key)] = value
	}

	orderbooks := &OrderbookModel{
		Orderbooks:   make(map[string]*orderbook.OrderBook),
		db:           batchDB,
		allowedPairs: fixAllowedPairs,
		// BzzClient: client.NewClient(bzzURL),
		// Signer:    signer,
	}

	return orderbooks
}

func (m *OrderbookModel) GetOrderBook(pairName string) (*orderbook.OrderBook, error) {
	return m.getAndCreateIfNotExisted(pairName)
}

func (m *OrderbookModel) hasOrderBook(name string) bool {
	_, ok := m.Orderbooks[name]
	return ok
}

// func (m *OrderbookModel) getQuery(topic feed.Topic, address common.Address) *feed.Query {
// 	fd := &feed.Feed{
// 		Topic: topic,
// 		User:  address,
// 	}

// 	// time := time.Now().UnixNano()/int64(time.Millisecond) - Month
// 	epoch := lookup.Epoch{
// 		Level: 25,
// 		Time:  1538650124,
// 	}
// 	// epoch := lookup.NoClue

// 	// demo.LogInfo("hash:", "address", topic.Hex())

// 	return feed.NewQueryLatest(fd, epoch)
// }

// commit for all orderbooks
func (m *OrderbookModel) Commit() error {
	return m.db.Commit()
}

// func (m *OrderbookModel) GetOrdersByAddress(coin, signerAddress string) ([]OrderbookMsg, error) {
// 	// topic, _ := m.getTopic(coin)
// 	topic, _ := m.getTopic(coin)
// 	address := common.HexToAddress(signerAddress)
// 	lookupParams := m.getQuery(topic, address)
// 	return m.GetOrdersByQuery(lookupParams)
// }

// func (m *OrderbookModel) GetOrdersByQuery(lookupParams *feed.Query) (messages []OrderbookMsg, err error) {

// 	reader, err := m.BzzClient.QueryFeed(lookupParams, "")

// 	// reader, err := m.BzzClient.QueryFeed(nil, hash)

// 	if err != nil {
// 		return nil, fmt.Errorf("Error retrieving feed updates: %s", err)
// 	}
// 	defer reader.Close()
// 	databytes, err := ioutil.ReadAll(reader)

// 	if databytes == nil || err != nil {
// 		return nil, err
// 	}

// 	// try to decode
// 	err = rlp.DecodeBytes(databytes, &messages)
// 	// demo.LogInfo("Data bytes", "messages", messages)
// 	return messages, err
// }
func (m *OrderbookModel) getAndCreateIfNotExisted(pairName string) (*orderbook.OrderBook, error) {

	name := strings.ToLower(pairName)

	if !m.hasOrderBook(name) {
		// check allow pair
		if _, ok := m.allowedPairs[name]; !ok {
			return nil, fmt.Errorf("Orderbook not found for pair :%s", pairName)
		}

		// then create one
		ob := orderbook.NewOrderBook(name, m.db)
		if ob != nil {
			ob.Restore()
			m.Orderbooks[name] = ob
		}
	}

	// return from map
	return m.Orderbooks[name], nil
}
func (m *OrderbookModel) GetOrder(pairName, orderID string) *orderbook.Order {
	ob, _ := m.getAndCreateIfNotExisted(pairName)
	if ob == nil {
		return nil
	}
	key := orderbook.GetKeyFromString(orderID)
	return ob.GetOrder(key)
}

func (m *OrderbookModel) ProcessOrder(orderbookMsg *OrderbookMsg) ([]map[string]string, map[string]string) {

	ob, _ := m.getAndCreateIfNotExisted(orderbookMsg.PairName)
	var trades []map[string]string
	var orderInBook map[string]string

	if ob != nil {
		// get map as general input, we can set format later to make sure there is no problem
		quote := orderbookMsg.ToQuote()
		trades, orderInBook = ob.ProcessOrder(quote, true)
	}
	// topic, _ := m.getTopic(orderbookMsg.Coin)
	// address := m.Signer.Address()

	// // demo.LogInfo("hash:", "address", topic.Hex())
	// lookupParams := m.getQuery(topic, address)

	// messages, err := m.GetOrdersByQuery(lookupParams)
	// var isNew = false
	// if messages == nil {
	// 	isNew = true
	// 	messages = []OrderbookMsg{*orderbookMsg}
	// } else {
	// 	// find item if found then append, else update
	// 	var found = false
	// 	for i, message := range messages {
	// 		if message.ID == orderbookMsg.ID {
	// 			found = true
	// 			messages[i] = *orderbookMsg
	// 			break
	// 		}
	// 	}
	// 	if !found {
	// 		messages = append(messages, *orderbookMsg)
	// 	}
	// }

	// databytes, err := rlp.EncodeToBytes(messages)
	// if err != nil {
	// 	return fmt.Errorf("Can not serialize data: %s", err)
	// }

	// if isNew {
	// 	createRequest := feed.NewFirstRequest(topic)

	// 	createRequest.SetData(databytes)
	// 	if err := createRequest.Sign(m.Signer); err != nil {
	// 		return fmt.Errorf("Error signing update: %s", err)
	// 	}

	// 	hashAddress, err := m.BzzClient.CreateFeedWithManifest(createRequest)
	// 	demo.LogInfo("Save to swarm", "address", hashAddress, "error", err)
	// 	return err
	// }

	// // other wise just update it
	// updateRequest, err := m.BzzClient.GetFeedRequest(lookupParams, "")
	// if err != nil {
	// 	return fmt.Errorf("Error retrieving update request template: %s", err)
	// }

	// updateRequest.SetData(databytes)
	// if err := updateRequest.Sign(m.Signer); err != nil {
	// 	return fmt.Errorf("Error signing update: %s", err)
	// }

	// if err = m.BzzClient.UpdateFeed(updateRequest); err != nil {
	// 	return fmt.Errorf("Error updating feed: %s", err)
	// }

	return trades, orderInBook

}

// func (m *OrderbookModel) UpdateData(coin, signerAddress string, epoch lookup.Epoch, hexData, hexSignature string) error {
// 	topic, _ := m.getTopic(coin)
// 	address := m.Signer.Address()

// 	// demo.LogInfo("hash:", "address", topic.Hex())
// 	lookupParams := m.getQuery(topic, address)
// 	request, _ := m.BzzClient.GetFeedRequest(lookupParams, "")
// 	request.Epoch = epoch
// 	data := common.Hex2Bytes(hexData)
// 	request.SetData(data)
// 	// request.Sign(m.Signer)
// 	correct := request.Signature
// 	var signature feed.Signature
// 	signaturebytes := common.Hex2Bytes(hexSignature)
// 	copy(signature[:], signaturebytes)
// 	request.Signature = &signature

// 	demo.LogInfo("Testing", "signature", fmt.Sprintf("%0x", signature), "correct", fmt.Sprintf("%0x", correct))
// 	if err := m.BzzClient.UpdateFeed(request); err != nil {
// 		return fmt.Errorf("Error updating feed: %s", err)
// 	}

// 	return nil
// }
