package node

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/sync"
)

type Server struct {
	SyncChecker           sync.Checker
	OptimisticModeFetcher blockchain.OptimisticModeFetcher
	BeaconDB              db.ReadOnlyDatabase
	PeersFetcher          p2p.PeersProvider
	PeerManager           p2p.PeerManager
	MetadataProvider      p2p.MetadataProvider
	GenesisTimeFetcher    blockchain.TimeFetcher
	HeadFetcher           blockchain.HeadFetcher
	SilaChainInfoFetcher  silaexec.ChainInfoFetcher
}
