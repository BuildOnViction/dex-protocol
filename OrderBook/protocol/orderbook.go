package protocol

import (
	"fmt"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	demo "../../common"
	"../orderbook"
	"github.com/ethereum/go-ethereum/swarm/api/client"
	"github.com/ethereum/go-ethereum/swarm/storage/feed"
	"github.com/ethereum/go-ethereum/swarm/storage/feed/lookup"
)

const (
	Day       = 60 * 60 * 24
	Year      = Day * 365
	Month     = Day * 30
	TopicName = "Token"
)

// OrderBookModel : singleton orderbook for testing
type OrderbookModel struct {
	Orderbook *orderbook.OrderBook
	BzzClient *client.Client
	Signer    *feed.GenericSigner
}

func NewModel(bzzURL string, signer *feed.GenericSigner) *OrderbookModel {
	demo.LogDebug("Creating model", "signerAddress", signer.Address().Hex())
	return &OrderbookModel{
		Orderbook: orderbook.NewOrderBook(),
		BzzClient: client.NewClient(bzzURL),
		Signer:    signer,
	}
}

func (m *OrderbookModel) getTopic(coin string) (feed.Topic, error) {
	return feed.NewTopic(TopicName, []byte(coin))
}

func (m *OrderbookModel) getQuery(topic feed.Topic, address common.Address) *feed.Query {
	fd := &feed.Feed{
		Topic: topic,
		User:  address,
	}

	// time := time.Now().UnixNano()/int64(time.Millisecond) - Month
	// epoch := lookup.Epoch{
	// 	Level: 13,
	// 	Time:  uint64(time),
	// }
	epoch := lookup.NoClue

	// demo.LogInfo("hash:", "address", topic.Hex())

	return feed.NewQueryLatest(fd, epoch)
}

func (m *OrderbookModel) GetOrdersByAddress(coin, signerAddress string) ([]OrderbookMsg, error) {
	// topic, _ := m.getTopic(coin)
	topic, _ := m.getTopic(coin)
	address := common.HexToAddress(signerAddress)
	lookupParams := m.getQuery(topic, address)
	return m.GetOrdersByQuery(lookupParams)
}

func (m *OrderbookModel) GetOrdersByQuery(lookupParams *feed.Query) (messages []OrderbookMsg, err error) {

	reader, err := m.BzzClient.QueryFeed(lookupParams, "")

	// reader, err := m.BzzClient.QueryFeed(nil, hash)

	if err != nil {
		return nil, fmt.Errorf("Error retrieving feed updates: %s", err)
	}
	defer reader.Close()
	databytes, err := ioutil.ReadAll(reader)

	if databytes == nil || err != nil {
		return nil, err
	}

	// try to decode
	err = rlp.DecodeBytes(databytes, &messages)
	// demo.LogInfo("Data bytes", "messages", messages)
	return messages, err
}

func (m *OrderbookModel) ProcessOrder(orderbookMsg *OrderbookMsg) error {

	topic, _ := m.getTopic(orderbookMsg.Coin)
	address := m.Signer.Address()

	// demo.LogInfo("hash:", "address", topic.Hex())
	lookupParams := m.getQuery(topic, address)

	messages, err := m.GetOrdersByQuery(lookupParams)
	var isNew = false
	if messages == nil {
		isNew = true
		messages = []OrderbookMsg{*orderbookMsg}
	} else {
		// find item if found then append, else update
		var found = false
		for i, message := range messages {
			if message.ID == orderbookMsg.ID {
				found = true
				messages[i] = *orderbookMsg
				break
			}
		}
		if !found {
			messages = append(messages, *orderbookMsg)
		}
	}

	databytes, err := rlp.EncodeToBytes(messages)
	if err != nil {
		return fmt.Errorf("Can not serialize data: %s", err)
	}

	if isNew {
		createRequest := feed.NewFirstRequest(topic)

		createRequest.SetData(databytes)
		if err := createRequest.Sign(m.Signer); err != nil {
			return fmt.Errorf("Error signing update: %s", err)
		}

		hashAddress, err := m.BzzClient.CreateFeedWithManifest(createRequest)
		demo.LogInfo("Save to swarm", "address", hashAddress, "error", err)
		return err
	}

	// other wise just update it
	updateRequest, err := m.BzzClient.GetFeedRequest(lookupParams, "")
	if err != nil {
		return fmt.Errorf("Error retrieving update request template: %s", err)
	}

	updateRequest.SetData(databytes)
	if err := updateRequest.Sign(m.Signer); err != nil {
		return fmt.Errorf("Error signing update: %s", err)
	}

	if err = m.BzzClient.UpdateFeed(updateRequest); err != nil {
		return fmt.Errorf("Error updating feed: %s", err)
	}

	return nil

}
