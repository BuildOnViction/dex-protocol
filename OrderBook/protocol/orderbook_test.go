package protocol

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/storage/feed/lookup"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/swarm/storage/feed"
)

func TestRLPEncode(t *testing.T) {
	msg := []OrderbookMsg{{
		Coin:      "Tomo",
		ID:        "2",
		Price:     "100",
		Quantity:  "10",
		Side:      "ask",
		Timestamp: 1538650124,
		TradeID:   "1",
		Type:      "limit",
	}}

	data, _ := rlp.EncodeToBytes(msg)

	topic, _ := feed.NewTopic(TopicName, []byte("Tomo"))
	request := new(feed.Request)

	// get the current time

	request.Epoch = lookup.Epoch{
		Time:  msg[0].Timestamp,
		Level: 25,
	}
	request.Feed.Topic = topic
	request.Header.Version = 0
	request.Feed.User = common.HexToAddress("0x28074f8D0fD78629CD59290Cac185611a8d60109")
	request.SetData(data)
	digest, _ := request.GetDigest()
	dataHex := fmt.Sprintf("0x%s", common.Bytes2Hex(data))

	t.Logf("Data: %s", dataHex)
	t.Logf("Topic hex: %s", topic.Hex())
	t.Logf("Message digest: %s", digest.Hex())

	if topic.Hex() != "0x0000060a6e000000000000000000000000000000000000000000000000000000" {
		t.Fatalf("topic hex is not correct")
	}

	if dataHex != "0xdedd845bb5f00c856c696d69748361736b8231308331303084546f6d6f3231" {
		t.Fatalf("data hex is not correct")
	}

	if digest.Hex() != "0xb40bfc0e66053d4b0525bfe683ca0e85a79ce96d228691ac342994bbaaa0ac97" {
		t.Fatalf("digest is not correct")
	}

	keyBytes := common.Hex2Bytes("3411b45169aa5a8312e51357db68621031020dcf46011d7431db1bbb6d3922ce")
	privkey, _ := crypto.ToECDSA(keyBytes)
	signer := feed.NewGenericSigner(privkey)
	signature, _ := signer.Sign(digest)

	t.Logf("Signature: %0x", signature)

}
