package validator

import (
	"math/big"
	"testing"

	mock "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache/depositsnapshot"
	mockSila "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec/testing"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/container/trie"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func TestShouldFallback(t *testing.T) {
	tests := []struct {
		name            string
		totalDepCount   uint64
		unFinalizedDeps uint64
		want            bool
	}{
		{
			name:            "0 dep count",
			totalDepCount:   0,
			unFinalizedDeps: 100,
			want:            false,
		},
		{
			name:            "0 unfinalized count",
			totalDepCount:   100,
			unFinalizedDeps: 0,
			want:            false,
		},
		{
			name:            "equal number of deposits and non finalized deposits",
			totalDepCount:   1000,
			unFinalizedDeps: 1000,
			want:            true,
		},
		{
			name:            "large number of non finalized deposits",
			totalDepCount:   300000,
			unFinalizedDeps: 100000,
			want:            true,
		},
		{
			name:            "small number of non finalized deposits",
			totalDepCount:   300000,
			unFinalizedDeps: 2000,
			want:            false,
		},
		{
			name:            "unfinalized deposits beyond threshold",
			totalDepCount:   300000,
			unFinalizedDeps: 10000,
			want:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRebuildTrie(tt.totalDepCount, tt.unFinalizedDeps); got != tt.want {
				t.Errorf("shouldRebuildTrie() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposer_PendingDeposits_Electra(t *testing.T) {
	// Electra continues to pack deposits while the state silaexecdeposit index is less than the silaexecdepositIndexLimit
	ctx := t.Context()

	height := big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance))
	newHeight := big.NewInt(height.Int64() + 11000)
	p := &mockSila.Chain{
		LatestBlockNumber: height,
		HashesByHeight: map[int][]byte{
			int(height.Int64()):    []byte("0x0"),
			int(newHeight.Int64()): []byte("0x1"),
		},
	}

	var votes []*silapb.SilaData

	vote := &silapb.SilaData{
		BlockHash:    bytesutil.PadTo([]byte("0x1"), 32),
		DepositRoot:  make([]byte, 32),
		DepositCount: 7,
	}
	period := uint64(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().EpochsPerSilaExecutionVotingPeriod)))
	for i := 0; i <= int(period/2); i++ {
		votes = append(votes, vote)
	}

	beaconState, err := state_native.InitializeFromProtoElectra(&silapb.BeaconStateElectra{
		SilaData: &silapb.SilaData{
			BlockHash:    []byte("0x0"),
			DepositRoot:  make([]byte, 32),
			DepositCount: 5,
		},
		SilaExecutionDepositIndex: 1,
		DepositRequestsStartIndex: 7,
		SilaDataVotes:             votes,
	})
	require.NoError(t, err)
	blk := util.NewBeaconBlockElectra()
	blk.Block.Slot = beaconState.Slot()

	blkRoot, err := blk.HashTreeRoot()
	require.NoError(t, err)

	var mockSig [96]byte
	var mockCreds [32]byte

	// Using the merkleTreeIndex as the block number for this test...
	readyDeposits := []*silapb.DepositContainer{
		{
			Index:           0,
			SilaBlockHeight: 8,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("a"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index:           1,
			SilaBlockHeight: 14,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("b"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	recentDeposits := []*silapb.DepositContainer{
		{
			Index:           2,
			SilaBlockHeight: 5000,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("c"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index:           3,
			SilaBlockHeight: 6000,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("d"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	depositTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	require.NoError(t, err, "Could not setup deposit trie")
	for _, dp := range append(readyDeposits, recentDeposits...) {
		depositHash, err := dp.Deposit.Data.HashTreeRoot()
		require.NoError(t, err, "Unable to determine hashed value of deposit")

		assert.NoError(t, depositTrie.Insert(depositHash[:], int(dp.Index)))
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		assert.NoError(t, depositCache.InsertDeposit(ctx, dp.Deposit, dp.SilaBlockHeight, dp.Index, root))
	}
	for _, dp := range recentDeposits {
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		depositCache.InsertPendingDeposit(ctx, dp.Deposit, dp.SilaBlockHeight, dp.Index, root)
	}

	bs := &Server{
		ChainStartFetcher:      p,
		SilaChainInfoFetcher:   p,
		SilaBlockFetcher:       p,
		DepositFetcher:         depositCache,
		PendingDepositsFetcher: depositCache,
		BlockReceiver:          &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		HeadFetcher:            &mock.ChainService{State: beaconState, Root: blkRoot[:]},
	}

	deposits, err := bs.deposits(ctx, beaconState, &silapb.SilaData{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(deposits), "Received unexpected list of deposits")

	// It should also return the recent deposits after their follow window.
	p.LatestBlockNumber = big.NewInt(0).Add(p.LatestBlockNumber, big.NewInt(10000))
	// we should get our pending deposits once this vote pushes the vote tally to include
	// the updated silaexec data.
	deposits, err = bs.deposits(ctx, beaconState, vote)
	require.NoError(t, err)
	assert.Equal(t, len(recentDeposits), len(deposits), "Received unexpected number of pending deposits")

	require.NoError(t, beaconState.SetDepositRequestsStartIndex(0)) // set it to 0 so it's less than SilaExecutionDepositIndex
	deposits, err = bs.deposits(ctx, beaconState, vote)
	require.NoError(t, err)
	assert.Equal(t, 0, len(deposits), "Received unexpected number of pending deposits")

}
