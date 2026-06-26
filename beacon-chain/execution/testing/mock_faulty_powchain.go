package testing

import (
	"context"
	"math/big"

	"github.com/sila-chain/Sila-Consensus-Core/v7/async/event"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/execution/types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila/common"
	"github.com/pkg/errors"
)

// FaultyExecutionChain defines an incorrectly functioning powchain service.
type FaultyExecutionChain struct {
	ChainFeed      *event.Feed
	HashesByHeight map[int][]byte
}

// SilaConsensusGenesisPowchainInfo --
func (*FaultyExecutionChain) SilaConsensusGenesisPowchainInfo() (uint64, *big.Int) {
	return 0, big.NewInt(0)
}

// BlockExists --
func (f *FaultyExecutionChain) BlockExists(context.Context, common.Hash) (bool, *big.Int, error) {
	if f.HashesByHeight == nil {
		return false, big.NewInt(1), errors.New("failed")
	}

	return true, big.NewInt(1), nil
}

// BlockHashByHeight --
func (*FaultyExecutionChain) BlockHashByHeight(context.Context, *big.Int) (common.Hash, error) {
	return [32]byte{}, errors.New("failed")
}

// BlockTimeByHeight --
func (*FaultyExecutionChain) BlockTimeByHeight(context.Context, *big.Int) (uint64, error) {
	return 0, errors.New("failed")
}

// BlockByTimestamp --
func (*FaultyExecutionChain) BlockByTimestamp(context.Context, uint64) (*types.HeaderInfo, error) {
	return &types.HeaderInfo{Number: big.NewInt(0)}, nil
}

// ChainStartSilaData --
func (*FaultyExecutionChain) ChainStartSilaData() *silapb.SilaData {
	return &silapb.SilaData{}
}

// PreGenesisState --
func (*FaultyExecutionChain) PreGenesisState() state.BeaconState {
	s, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	if err != nil {
		panic("could not initialize state") // lint:nopanic -- test code.
	}
	return s
}

// ClearPreGenesisData --
func (*FaultyExecutionChain) ClearPreGenesisData() {
	// no-op
}

// IsConnectedToSilaExecution --
func (*FaultyExecutionChain) IsConnectedToSilaExecution() bool {
	return true
}
