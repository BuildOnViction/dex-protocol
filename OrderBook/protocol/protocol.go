package protocol

import (
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"

	demo "github.com/tomochain/orderbook/common"
	"github.com/tomochain/orderbook/orderbook"
)

var (
	OrderbookProtocol = &protocols.Spec{
		Name:       demo.FooProtocolName,
		Version:    demo.FooProtocolVersion,
		MaxMsgSize: demo.FooProtocolMaxMsgSize,
		Messages: []interface{}{
			&OrderbookHandshake{}, &OrderbookMsg{},
		},
	}
)

type OrderbookMsg struct {
	PairName  string
	ID        string
	Price     string
	Quantity  string
	Side      string
	Timestamp uint64
	TradeID   string
	Type      string
}

func (msg *OrderbookMsg) ToQuote() map[string]string {
	quote := make(map[string]string)
	quote["timestamp"] = strconv.FormatUint(msg.Timestamp, 10)
	quote["type"] = msg.Type
	quote["side"] = msg.Side
	quote["quantity"] = msg.Quantity
	quote["price"] = msg.Price
	quote["trade_id"] = msg.TradeID
	quote["pairName"] = msg.PairName
	// if insert id is not used, just for update
	quote["id"] = msg.ID
	return quote
}

func NewOrderbookMsg(quote map[string]string) (*OrderbookMsg, error) {
	timestamp, err := strconv.ParseUint(quote["timestamp"], 10, 64)
	return &OrderbookMsg{
		Timestamp: timestamp,
		Type:      quote["type"],
		Side:      quote["side"],
		Quantity:  quote["quantity"],
		Price:     quote["price"],
		TradeID:   quote["trade_id"],
		PairName:  quote["pairName"],
		ID:        quote["id"],
	}, err
}

type OrderbookHandshake struct {
	Nick string
	V    uint
}

// the protocols abstraction enables use of an external handler function
type OrderbookHandler struct {
	Engine *orderbook.Engine
	Peer   *protocols.Peer
	InC    <-chan interface{}
	QuitC  <-chan struct{}
}

// we will receive message in handle
func (orderbookHandler *OrderbookHandler) handle(msg interface{}) error {

	// we got message or handshake
	message, ok := msg.(*OrderbookMsg)

	// switch message := x.(type) {
	// case *OrderbookMsg:

	// }

	if ok {

		demo.LogDebug("Received orderbook", "orderbook", message, "peer", orderbookHandler.Peer)

		// add Order
		payload := message.ToQuote()
		demo.LogInfo("-> Add order", "payload", payload)

		trades, orderInBook := orderbookHandler.Engine.ProcessOrder(payload)
		demo.LogInfo("Orderbook result", "Trade", trades, "OrderInBook", orderInBook)
		return nil
	}

	orderbookhs, ok := msg.(*OrderbookHandshake)
	if ok {
		demo.LogDebug("Processing handshake", "from", orderbookhs.Nick, "version", orderbookhs.V)
		// now protocol is ok, we can inject channel to receive message
		go func() {
			for {
				select {
				case payload := <-orderbookHandler.InC:
					// demo.LogInfo("Internal received", "payload", payload)
					inmsg, ok := payload.(*OrderbookMsg)
					if ok {
						// maybe we have to use map[]chan
						// databytes, err := rlp.EncodeToBytes(inmsg)
						// databytes, err := json.Marshal(inmsg)

						demo.LogDebug("Sending orderbook", "orderbook", inmsg)
						orderbookHandler.Peer.Send(inmsg)

					}

				// send quit command, break this loop
				case <-orderbookHandler.QuitC:
					break
				}
			}
		}()
		return nil
	}

	return fmt.Errorf("Invalid pssorderbook protocol message")

}

// create the protocol with the protocols extension
func NewProtocol(inC <-chan interface{}, quitC <-chan struct{}, orderbookEngine *orderbook.Engine) *p2p.Protocol {
	return &p2p.Protocol{
		Name:    "Orderbook",
		Version: 42,
		Length:  1,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			demo.LogWarn("running", "peer", p)
			// create the enhanced peer, it will wrap p2p.Send with code from Message Spec
			pp := protocols.NewPeer(p, rw, OrderbookProtocol)

			// send the message, then handle it to make sure protocol success
			go func() {
				outmsg := &OrderbookHandshake{
					V: 42,
					// shortened hex string for terminal logging
					Nick: p.Name(),
				}
				err := pp.Send(outmsg)
				if err != nil {
					demo.LogError("Send p2p message fail", "err", err)
				}
				demo.LogInfo("Sending handshake", "peer", p, "handshake", outmsg)
			}()

			// protocols abstraction provides a separate blocking run loop for the peer
			// when this returns, the protocol will be terminated
			run := &OrderbookHandler{
				Engine: orderbookEngine,
				Peer:   pp,
				// assign channel
				InC:   inC,
				QuitC: quitC,
			}
			err := pp.Run(run.handle)
			return err
		},
	}
}
