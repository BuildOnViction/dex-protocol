package protocol

import (
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
)

// the service we want to offer on the node
// it must implement the node.Service interface
type OrderbookService struct {
	V     int
	Model *OrderbookModel
}

// APIs : api service
// specify API structs that carry the methods we want to use
func (service *OrderbookService) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "orderbook",
			Version:   "0.42",
			Service:   NewOrderbookAPI(service.V, service.Model),
			Public:    true,
		},
	}
}

// these are needed to satisfy the node.Service interface
// in this example they do nothing
func (service *OrderbookService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

func (service *OrderbookService) Start(srv *p2p.Server) error {
	return nil
}

func (service *OrderbookService) Stop() error {
	return nil
}

// wrapper function for servicenode to start the service
func NewService(orderbookModel *OrderbookModel) func(ctx *node.ServiceContext) (node.Service, error) {
	return func(ctx *node.ServiceContext) (node.Service, error) {
		return &OrderbookService{
			V:     42,
			Model: orderbookModel,
		}, nil
	}
}
