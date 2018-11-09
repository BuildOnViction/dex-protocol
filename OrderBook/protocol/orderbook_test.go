package protocol

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/storage/feed/lookup"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/swarm/storage/feed"
)

func putint(b []byte, i uint64) (size int) {
	switch {
	case i < (1 << 8):
		b[0] = byte(i)
		return 1
	case i < (1 << 16):
		b[0] = byte(i >> 8)
		b[1] = byte(i)
		return 2
	case i < (1 << 24):
		b[0] = byte(i >> 16)
		b[1] = byte(i >> 8)
		b[2] = byte(i)
		return 3
	case i < (1 << 32):
		b[0] = byte(i >> 24)
		b[1] = byte(i >> 16)
		b[2] = byte(i >> 8)
		b[3] = byte(i)
		return 4
	case i < (1 << 40):
		b[0] = byte(i >> 32)
		b[1] = byte(i >> 24)
		b[2] = byte(i >> 16)
		b[3] = byte(i >> 8)
		b[4] = byte(i)
		return 5
	case i < (1 << 48):
		b[0] = byte(i >> 40)
		b[1] = byte(i >> 32)
		b[2] = byte(i >> 24)
		b[3] = byte(i >> 16)
		b[4] = byte(i >> 8)
		b[5] = byte(i)
		return 6
	case i < (1 << 56):
		b[0] = byte(i >> 48)
		b[1] = byte(i >> 40)
		b[2] = byte(i >> 32)
		b[3] = byte(i >> 24)
		b[4] = byte(i >> 16)
		b[5] = byte(i >> 8)
		b[6] = byte(i)
		return 7
	default:
		b[0] = byte(i >> 56)
		b[1] = byte(i >> 48)
		b[2] = byte(i >> 40)
		b[3] = byte(i >> 32)
		b[4] = byte(i >> 24)
		b[5] = byte(i >> 16)
		b[6] = byte(i >> 8)
		b[7] = byte(i)
		return 8
	}
}

func TestRLP(t *testing.T) {
	msg := []OrderbookMsg{
		{
			ID:        "1",
			Coin:      "Tomo",
			Price:     "100",
			Quantity:  "20",
			Side:      "ask",
			Timestamp: 1538650124,
			TradeID:   "1",
			Type:      "limit",
		},
		{
			ID:        "2",
			Coin:      "Tomo",
			Price:     "100",
			Quantity:  "20",
			Side:      "ask",
			Timestamp: 1538650124,
			TradeID:   "1",
			Type:      "limit",
		}}
	val := reflect.Indirect(reflect.ValueOf(&msg[0]))
	for i := 0; i < val.Type().NumField(); i++ {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s", val.Type().Field(i).Name)
	}
	fmt.Println()

	data, _ := rlp.EncodeToBytes(msg)
	t.Log(data)
}

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

	keyBytes := common.Hex2Bytes("3411b45169aa5a8312e51357db68621031020dcf46011d7431db1bbb6d3922ce")
	privkey, _ := crypto.ToECDSA(keyBytes)
	signer := feed.NewGenericSigner(privkey)
	signature, _ := signer.Sign(digest)

	t.Logf("Signature: %0x", signature)

	if topic.Hex() != "0x0000060a6e000000000000000000000000000000000000000000000000000000" {
		t.Fatalf("topic hex is not correct")
	}

	if dataHex != "0xdedd845bb5f00c856c696d69748361736b8231308331303084546f6d6f3231" {
		t.Fatalf("data hex is not correct")
	}

	if digest.Hex() != "0xb40bfc0e66053d4b0525bfe683ca0e85a79ce96d228691ac342994bbaaa0ac97" {
		t.Fatalf("digest is not correct")
	}

}
