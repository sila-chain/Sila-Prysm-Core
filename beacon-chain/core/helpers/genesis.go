package helpers

import (
	"fmt"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/container/trie"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// UpdateGenesisSilaData updates silaexec data for genesis state.
func UpdateGenesisSilaData(state state.BeaconState, deposits []*silapb.Deposit, silaexecData *silapb.SilaData) (state.BeaconState, error) {
	if silaexecData == nil {
		return nil, errors.New("no silaData provided for genesis state")
	}

	leaves := make([][]byte, 0, len(deposits))
	for _, deposit := range deposits {
		if deposit == nil || deposit.Data == nil {
			return nil, fmt.Errorf("nil deposit or deposit with nil data cannot be processed: %v", deposit)
		}
		hash, err := deposit.Data.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, hash[:])
	}
	var t *trie.SparseMerkleTrie
	var err error
	if len(leaves) > 0 {
		t, err = trie.GenerateTrieFromItems(leaves, params.BeaconConfig().SilaDepositTreeDepth)
		if err != nil {
			return nil, err
		}
	} else {
		t, err = trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
		if err != nil {
			return nil, err
		}
	}

	depositRoot, err := t.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	silaexecData.DepositRoot = depositRoot[:]
	err = state.SetSilaData(silaexecData)
	if err != nil {
		return nil, err
	}
	return state, nil
}
