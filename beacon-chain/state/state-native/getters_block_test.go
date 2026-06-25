package state_native

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	testtmpl "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func TestBeaconState_LatestBlockHeader_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateLatestBlockHeader(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoPhase0(&silapb.BeaconState{})
		},
		func(BH *silapb.BeaconBlockHeader) (state.BeaconState, error) {
			return InitializeFromProtoPhase0(&silapb.BeaconState{LatestBlockHeader: BH})
		},
	)
}

func TestBeaconState_LatestBlockHeader_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateLatestBlockHeader(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoAltair(&silapb.BeaconStateAltair{})
		},
		func(BH *silapb.BeaconBlockHeader) (state.BeaconState, error) {
			return InitializeFromProtoAltair(&silapb.BeaconStateAltair{LatestBlockHeader: BH})
		},
	)
}

func TestBeaconState_LatestBlockHeader_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateLatestBlockHeader(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoBellatrix(&silapb.BeaconStateBellatrix{})
		},
		func(BH *silapb.BeaconBlockHeader) (state.BeaconState, error) {
			return InitializeFromProtoBellatrix(&silapb.BeaconStateBellatrix{LatestBlockHeader: BH})
		},
	)
}

func TestBeaconState_LatestBlockHeader_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateLatestBlockHeader(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoCapella(&silapb.BeaconStateCapella{})
		},
		func(BH *silapb.BeaconBlockHeader) (state.BeaconState, error) {
			return InitializeFromProtoCapella(&silapb.BeaconStateCapella{LatestBlockHeader: BH})
		},
	)
}

func TestBeaconState_LatestBlockHeader_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStateLatestBlockHeader(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoDeneb(&silapb.BeaconStateDeneb{})
		},
		func(BH *silapb.BeaconBlockHeader) (state.BeaconState, error) {
			return InitializeFromProtoDeneb(&silapb.BeaconStateDeneb{LatestBlockHeader: BH})
		},
	)
}

func TestBeaconState_BlockRoots_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootsNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoPhase0(&silapb.BeaconState{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProtoPhase0(&silapb.BeaconState{BlockRoots: BR})
		},
	)
}

func TestBeaconState_BlockRoots_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootsNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoAltair(&silapb.BeaconStateAltair{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProtoAltair(&silapb.BeaconStateAltair{BlockRoots: BR})
		},
	)
}

func TestBeaconState_BlockRoots_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootsNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoBellatrix(&silapb.BeaconStateBellatrix{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProtoBellatrix(&silapb.BeaconStateBellatrix{BlockRoots: BR})
		},
	)
}

func TestBeaconState_BlockRoots_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootsNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoCapella(&silapb.BeaconStateCapella{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProtoCapella(&silapb.BeaconStateCapella{BlockRoots: BR})
		},
	)
}

func TestBeaconState_BlockRoots_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootsNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoDeneb(&silapb.BeaconStateDeneb{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProtoDeneb(&silapb.BeaconStateDeneb{BlockRoots: BR})
		},
	)
}

func TestBeaconState_BlockRootAtIndex_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootAtIndexNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoPhase0(&silapb.BeaconState{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProtoPhase0(&silapb.BeaconState{BlockRoots: BR})
		},
	)
}

func TestBeaconState_BlockRootAtIndex_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootAtIndexNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoAltair(&silapb.BeaconStateAltair{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProtoAltair(&silapb.BeaconStateAltair{BlockRoots: BR})
		},
	)
}

func TestBeaconState_BlockRootAtIndex_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootAtIndexNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoBellatrix(&silapb.BeaconStateBellatrix{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProtoBellatrix(&silapb.BeaconStateBellatrix{BlockRoots: BR})
		},
	)
}

func TestBeaconState_BlockRootAtIndex_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootAtIndexNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoCapella(&silapb.BeaconStateCapella{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProtoCapella(&silapb.BeaconStateCapella{BlockRoots: BR})
		},
	)
}

func TestBeaconState_BlockRootAtIndex_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStateBlockRootAtIndexNative(
		t,
		func() (state.BeaconState, error) {
			return InitializeFromProtoDeneb(&silapb.BeaconStateDeneb{})
		},
		func(BR [][]byte) (state.BeaconState, error) {
			return InitializeFromProtoDeneb(&silapb.BeaconStateDeneb{BlockRoots: BR})
		},
	)
}

func TestBeaconState_ProposerDependentRoot(t *testing.T) {
	slotsPerEpoch := uint64(params.BeaconConfig().SlotsPerEpoch)

	t.Run("epoch < 2 returns sentinel", func(t *testing.T) {
		s, err := InitializeFromProtoPhase0(&silapb.BeaconState{Slot: 1})
		require.NoError(t, err)
		_, err = s.ProposerDependentRoot(primitives.Slot(slotsPerEpoch - 1))
		require.ErrorIs(t, err, ErrProposerDependentRootUnderflow)
	})

	t.Run("happy path returns block_roots[epoch_start(epoch-1)-1]", func(t *testing.T) {
		var blockRoots [][]byte
		for i := uint64(0); i < uint64(params.BeaconConfig().SlotsPerHistoricalRoot); i++ {
			blockRoots = append(blockRoots, []byte{byte(i)})
		}
		// slot in epoch 2 → boundary = epoch_start(1) = SlotsPerEpoch → expect block_roots[SlotsPerEpoch-1].
		proposalSlot := primitives.Slot(2 * slotsPerEpoch)
		s, err := InitializeFromProtoPhase0(&silapb.BeaconState{
			BlockRoots: blockRoots,
			Slot:       primitives.Slot(3 * slotsPerEpoch),
		})
		require.NoError(t, err)

		got, err := s.ProposerDependentRoot(proposalSlot)
		require.NoError(t, err)
		var expected [32]byte
		expected[0] = byte(slotsPerEpoch - 1)
		assert.DeepEqual(t, expected, got)
	})
}
