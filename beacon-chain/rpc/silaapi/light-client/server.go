package lightclient

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain"
	lightClient "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/light-client"
)

type Server struct {
	LCStore     *lightClient.Store
	HeadFetcher blockchain.HeadFetcher
}
