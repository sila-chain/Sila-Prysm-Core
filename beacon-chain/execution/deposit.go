package execution

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/altair"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila/common"
	"github.com/pkg/errors"
)

// DepositContractAddress returns the deposit contract address for the given chain.
func DepositContractAddress() (string, error) {
	address := params.BeaconConfig().DepositContractAddress
	if address == "" {
		return "", errors.New("valid deposit contract is required")
	}

	if !common.IsHexAddress(address) {
		return "", errors.New("invalid deposit contract address given: " + address)
	}
	return address, nil
}

func (s *Service) processDeposit(ctx context.Context, eth1Data *silapb.Eth1Data, deposit *silapb.Deposit) error {
	var err error
	if err := s.preGenesisState.SetEth1Data(eth1Data); err != nil {
		return err
	}
	// preGenesisState is always a genesis state ( phase 0 ) and so state version does not need to be checked here for post electra deposit processing
	beaconState, err := altair.ProcessPreGenesisDeposits(ctx, s.preGenesisState, []*silapb.Deposit{deposit})
	if err != nil {
		return errors.Wrap(err, "could not process pre-genesis deposits")
	}
	if beaconState != nil && !beaconState.IsNil() {
		s.preGenesisState = beaconState
	}
	return nil
}
