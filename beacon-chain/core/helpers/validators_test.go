package helpers_test

import (
	"errors"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/time"
	forkchoicetypes "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/forkchoice/types"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/hash"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func TestIsActiveValidator_OK(t *testing.T) {
	tests := []struct {
		a primitives.Epoch
		b bool
	}{
		{a: 0, b: false},
		{a: 10, b: true},
		{a: 100, b: false},
		{a: 1000, b: false},
		{a: 64, b: true},
	}
	for _, test := range tests {
		validator := &silapb.Validator{ActivationEpoch: 10, ExitEpoch: 100}
		assert.Equal(t, test.b, helpers.IsActiveValidator(validator, test.a), "IsActiveValidator(%d)", test.a)
	}
}

func TestIsActiveValidatorUsingTrie_OK(t *testing.T) {
	tests := []struct {
		a primitives.Epoch
		b bool
	}{
		{a: 0, b: false},
		{a: 10, b: true},
		{a: 100, b: false},
		{a: 1000, b: false},
		{a: 64, b: true},
	}
	val := &silapb.Validator{ActivationEpoch: 10, ExitEpoch: 100}
	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{Validators: []*silapb.Validator{val}})
	require.NoError(t, err)
	for _, test := range tests {
		readOnlyVal, err := beaconState.ValidatorAtIndexReadOnly(0)
		require.NoError(t, err)
		assert.Equal(t, test.b, helpers.IsActiveValidatorUsingTrie(readOnlyVal, test.a), "IsActiveValidatorUsingTrie(%d)", test.a)
	}
}

func TestIsActiveNonSlashedValidatorUsingTrie_OK(t *testing.T) {
	tests := []struct {
		a primitives.Epoch
		s bool
		b bool
	}{
		{a: 0, s: false, b: false},
		{a: 10, s: false, b: true},
		{a: 100, s: false, b: false},
		{a: 1000, s: false, b: false},
		{a: 64, s: false, b: true},
		{a: 0, s: true, b: false},
		{a: 10, s: true, b: false},
		{a: 100, s: true, b: false},
		{a: 1000, s: true, b: false},
		{a: 64, s: true, b: false},
	}
	for _, test := range tests {
		val := &silapb.Validator{ActivationEpoch: 10, ExitEpoch: 100}
		val.Slashed = test.s
		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{Validators: []*silapb.Validator{val}})
		require.NoError(t, err)
		readOnlyVal, err := beaconState.ValidatorAtIndexReadOnly(0)
		require.NoError(t, err)
		assert.Equal(t, test.b, helpers.IsActiveNonSlashedValidatorUsingTrie(readOnlyVal, test.a), "IsActiveNonSlashedValidatorUsingTrie(%d)", test.a)
	}
}

func TestIsSlashableValidator_OK(t *testing.T) {
	tests := []struct {
		name      string
		validator *silapb.Validator
		epoch     primitives.Epoch
		slashable bool
	}{
		{
			name: "Unset withdrawable, slashable",
			validator: &silapb.Validator{
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			},
			epoch:     0,
			slashable: true,
		},
		{
			name: "before withdrawable, slashable",
			validator: &silapb.Validator{
				WithdrawableEpoch: 5,
			},
			epoch:     3,
			slashable: true,
		},
		{
			name: "inactive, not slashable",
			validator: &silapb.Validator{
				ActivationEpoch:   5,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			},
			epoch:     2,
			slashable: false,
		},
		{
			name: "after withdrawable, not slashable",
			validator: &silapb.Validator{
				WithdrawableEpoch: 3,
			},
			epoch:     3,
			slashable: false,
		},
		{
			name: "slashed and withdrawable, not slashable",
			validator: &silapb.Validator{
				Slashed:           true,
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: 1,
			},
			epoch:     2,
			slashable: false,
		},
		{
			name: "slashed, not slashable",
			validator: &silapb.Validator{
				Slashed:           true,
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			},
			epoch:     2,
			slashable: false,
		},
		{
			name: "inactive and slashed, not slashable",
			validator: &silapb.Validator{
				Slashed:           true,
				ActivationEpoch:   4,
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			},
			epoch:     2,
			slashable: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Run("without trie", func(t *testing.T) {
				slashableValidator := helpers.IsSlashableValidator(test.validator.ActivationEpoch,
					test.validator.WithdrawableEpoch, test.validator.Slashed, test.epoch)
				assert.Equal(t, test.slashable, slashableValidator, "Expected active validator slashable to be %t", test.slashable)
			})
			t.Run("with trie", func(t *testing.T) {
				beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{Validators: []*silapb.Validator{test.validator}})
				require.NoError(t, err)
				readOnlyVal, err := beaconState.ValidatorAtIndexReadOnly(0)
				require.NoError(t, err)
				slashableValidator := helpers.IsSlashableValidatorUsingTrie(readOnlyVal, test.epoch)
				assert.Equal(t, test.slashable, slashableValidator, "Expected active validator slashable to be %t", test.slashable)
			})
		})
	}
}

func TestBeaconProposerIndex_OK(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	c := params.BeaconConfig()
	c.MinGenesisActiveValidatorCount = 16384
	params.OverrideBeaconConfig(c)
	validators := make([]*silapb.Validator, params.BeaconConfig().MinGenesisActiveValidatorCount/8)
	for i := range validators {
		validators[i] = &silapb.Validator{
			ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		}
	}

	state, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		Validators:  validators,
		Slot:        0,
		RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
	})
	require.NoError(t, err)

	tests := []struct {
		slot  primitives.Slot
		index primitives.ValidatorIndex
	}{
		{
			slot:  1,
			index: 2039,
		},
		{
			slot:  5,
			index: 1895,
		},
		{
			slot:  19,
			index: 1947,
		},
		{
			slot:  30,
			index: 369,
		},
		{
			slot:  43,
			index: 464,
		},
	}

	for _, tt := range tests {
		helpers.ClearCache()

		require.NoError(t, state.SetSlot(tt.slot))
		result, err := helpers.BeaconProposerIndex(t.Context(), state)
		require.NoError(t, err, "Failed to get shard and committees at slot")
		assert.Equal(t, tt.index, result, "Result index was an unexpected value")
	}
}

func TestBeaconProposerIndex_BadState(t *testing.T) {
	helpers.ClearCache()

	params.SetupTestConfigCleanup(t)
	c := params.BeaconConfig()
	c.MinGenesisActiveValidatorCount = 16384
	params.OverrideBeaconConfig(c)
	validators := make([]*silapb.Validator, params.BeaconConfig().MinGenesisActiveValidatorCount/8)
	for i := range validators {
		validators[i] = &silapb.Validator{
			ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		}
	}
	roots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := uint64(0); i < uint64(params.BeaconConfig().SlotsPerHistoricalRoot); i++ {
		roots[i] = make([]byte, fieldparams.RootLength)
	}

	state, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		Validators:  validators,
		Slot:        0,
		RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
		BlockRoots:  roots,
		StateRoots:  roots,
	})
	require.NoError(t, err)
	// Set a very high slot, so that retrieved block root will be
	// non existent for the proposer cache.
	require.NoError(t, state.SetSlot(100))
	_, err = helpers.BeaconProposerIndex(t.Context(), state)
	require.NoError(t, err)
}

func TestComputeProposerIndex_Compatibility(t *testing.T) {
	helpers.ClearCache()

	validators := make([]*silapb.Validator, params.BeaconConfig().MinGenesisActiveValidatorCount)
	for i := range validators {
		validators[i] = &silapb.Validator{
			ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		}
	}

	state, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		Validators:  validators,
		RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
	})
	require.NoError(t, err)

	indices, err := helpers.ActiveValidatorIndices(t.Context(), state, 0)
	require.NoError(t, err)

	var proposerIndices []primitives.ValidatorIndex
	seed, err := helpers.Seed(state, 0, params.BeaconConfig().DomainBeaconProposer)
	require.NoError(t, err)
	for i := uint64(0); i < uint64(params.BeaconConfig().SlotsPerEpoch); i++ {
		seedWithSlot := append(seed[:], bytesutil.Bytes8(i)...)
		seedWithSlotHash := hash.Hash(seedWithSlot)
		index, err := helpers.ComputeProposerIndex(state, indices, seedWithSlotHash)
		require.NoError(t, err)
		proposerIndices = append(proposerIndices, index)
	}

	var wantedProposerIndices []primitives.ValidatorIndex
	seed, err = helpers.Seed(state, 0, params.BeaconConfig().DomainBeaconProposer)
	require.NoError(t, err)
	for i := uint64(0); i < uint64(params.BeaconConfig().SlotsPerEpoch); i++ {
		seedWithSlot := append(seed[:], bytesutil.Bytes8(i)...)
		seedWithSlotHash := hash.Hash(seedWithSlot)
		index, err := computeProposerIndexWithValidators(state.Validators(), indices, seedWithSlotHash)
		require.NoError(t, err)
		wantedProposerIndices = append(wantedProposerIndices, index)
	}
	assert.DeepEqual(t, wantedProposerIndices, proposerIndices, "Wanted proposer indices from ComputeProposerIndexWithValidators does not match")
}

func TestDelayedActivationExitEpoch_OK(t *testing.T) {
	helpers.ClearCache()

	epoch := primitives.Epoch(9999)
	wanted := epoch + 1 + params.BeaconConfig().MaxSeedLookahead
	assert.Equal(t, wanted, helpers.ActivationExitEpoch(epoch))
}

func TestActiveValidatorCount_Genesis(t *testing.T) {
	helpers.ClearCache()

	c := 1000
	validators := make([]*silapb.Validator, c)
	for i := range validators {
		validators[i] = &silapb.Validator{
			ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		}
	}
	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		Slot:        0,
		Validators:  validators,
		RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
	})
	require.NoError(t, err)

	// Preset cache to a bad count.
	seed, err := helpers.Seed(beaconState, 0, params.BeaconConfig().DomainBeaconAttester)
	require.NoError(t, err)
	require.NoError(t, helpers.CommitteeCache().AddCommitteeShuffledList(t.Context(), &cache.Committees{Seed: seed, ShuffledIndices: []primitives.ValidatorIndex{1, 2, 3}}))
	validatorCount, err := helpers.ActiveValidatorCount(t.Context(), beaconState, time.CurrentEpoch(beaconState))
	require.NoError(t, err)
	assert.Equal(t, uint64(c), validatorCount, "Did not get the correct validator count")
}

func TestChurnLimit_OK(t *testing.T) {
	tests := []struct {
		validatorCount int
		wantedChurn    uint64
	}{
		{validatorCount: 1000, wantedChurn: 4},
		{validatorCount: 100000, wantedChurn: 4},
		{validatorCount: 1000000, wantedChurn: 15 /* validatorCount/churnLimitQuotient */},
		{validatorCount: 2000000, wantedChurn: 30 /* validatorCount/churnLimitQuotient */},
	}
	for _, test := range tests {
		helpers.ClearCache()

		validators := make([]*silapb.Validator, test.validatorCount)
		for i := range validators {
			validators[i] = &silapb.Validator{
				ExitEpoch: params.BeaconConfig().FarFutureEpoch,
			}
		}

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot:        1,
			Validators:  validators,
			RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
		})
		require.NoError(t, err)
		validatorCount, err := helpers.ActiveValidatorCount(t.Context(), beaconState, time.CurrentEpoch(beaconState))
		require.NoError(t, err)
		resultChurn := helpers.ValidatorActivationChurnLimit(validatorCount)
		assert.Equal(t, test.wantedChurn, resultChurn, "ValidatorActivationChurnLimit(%d)", test.validatorCount)
	}
}

func TestChurnLimitDeneb_OK(t *testing.T) {
	tests := []struct {
		validatorCount int
		wantedChurn    uint64
	}{
		{1000, 4},
		{100000, 4},
		{1000000, params.BeaconConfig().MaxPerEpochActivationChurnLimit},
		{2000000, params.BeaconConfig().MaxPerEpochActivationChurnLimit},
	}
	for _, test := range tests {
		helpers.ClearCache()

		// Create validators
		validators := make([]*silapb.Validator, test.validatorCount)
		for i := range validators {
			validators[i] = &silapb.Validator{
				ExitEpoch: params.BeaconConfig().FarFutureEpoch,
			}
		}

		// Initialize beacon state
		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot:        1,
			Validators:  validators,
			RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
		})
		require.NoError(t, err)

		// Get active validator count
		validatorCount, err := helpers.ActiveValidatorCount(t.Context(), beaconState, time.CurrentEpoch(beaconState))
		require.NoError(t, err)

		// Test churn limit calculation
		resultChurn := helpers.ValidatorActivationChurnLimitDeneb(validatorCount)
		assert.Equal(t, test.wantedChurn, resultChurn)
	}
}

// Test basic functionality of ActiveValidatorIndices without caching. This test will need to be
// rewritten when releasing some cache flag.
func TestActiveValidatorIndices(t *testing.T) {
	//farFutureEpoch := params.BeaconConfig().FarFutureEpoch
	type args struct {
		state *silapb.BeaconState
		epoch primitives.Epoch
	}
	tests := []struct {
		name      string
		args      args
		want      []primitives.ValidatorIndex
		wantedErr string
	}{
		/*{
			name: "all_active_epoch_10",
			args: args{
				state: &silapb.BeaconState{
					RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
					Validators: []*silapb.Validator{
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
					},
				},
				epoch: 10,
			},
			want: []primitives.ValidatorIndex{0, 1, 2},
		},
		{
			name: "some_active_epoch_10",
			args: args{
				state: &silapb.BeaconState{
					RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
					Validators: []*silapb.Validator{
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       1,
						},
					},
				},
				epoch: 10,
			},
			want: []primitives.ValidatorIndex{0, 1},
		},
		{
			name: "some_active_with_recent_new_epoch_10",
			args: args{
				state: &silapb.BeaconState{
					RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
					Validators: []*silapb.Validator{
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       1,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
					},
				},
				epoch: 10,
			},
			want: []primitives.ValidatorIndex{0, 1, 3},
		},
		{
			name: "some_active_with_recent_new_epoch_10",
			args: args{
				state: &silapb.BeaconState{
					RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
					Validators: []*silapb.Validator{
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       1,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
					},
				},
				epoch: 10,
			},
			want: []primitives.ValidatorIndex{0, 1, 3},
		},
		{
			name: "some_active_with_recent_new_epoch_10",
			args: args{
				state: &silapb.BeaconState{
					RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
					Validators: []*silapb.Validator{
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       1,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
						{
							ActivationEpoch: 0,
							ExitEpoch:       farFutureEpoch,
						},
					},
				},
				epoch: 10,
			},
			want: []primitives.ValidatorIndex{0, 2, 3},
		},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.ClearCache()

			s, err := state_native.InitializeFromProtoPhase0(tt.args.state)
			require.NoError(t, err)
			got, err := helpers.ActiveValidatorIndices(t.Context(), s, tt.args.epoch)
			if tt.wantedErr != "" {
				assert.ErrorContains(t, tt.wantedErr, err)
				return
			}
			assert.DeepEqual(t, tt.want, got, "ActiveValidatorIndices()")
		})
	}
}

func TestComputeProposerIndex(t *testing.T) {
	seed := bytesutil.ToBytes32([]byte("seed"))
	type args struct {
		validators []*silapb.Validator
		indices    []primitives.ValidatorIndex
		seed       [32]byte
	}
	tests := []struct {
		name             string
		isElectraOrAbove bool
		args             args
		want             primitives.ValidatorIndex
		wantedErr        string
	}{
		{
			name: "all_active_indices",
			args: args{
				validators: []*silapb.Validator{
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
				},
				indices: []primitives.ValidatorIndex{0, 1, 2, 3, 4},
				seed:    seed,
			},
			want: 2,
		},
		{ // Regression test for https://github.com/sila-chain/sila/issues/4259.
			name: "1_active_index",
			args: args{
				validators: []*silapb.Validator{
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
				},
				indices: []primitives.ValidatorIndex{3},
				seed:    seed,
			},
			want: 3,
		},
		{
			name: "empty_active_indices",
			args: args{
				validators: []*silapb.Validator{
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
				},
				indices: []primitives.ValidatorIndex{},
				seed:    seed,
			},
			wantedErr: "empty active indices list",
		},
		{
			name: "active_indices_out_of_range",
			args: args{
				validators: []*silapb.Validator{
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
				},
				indices: []primitives.ValidatorIndex{100},
				seed:    seed,
			},
			wantedErr: "active index out of range",
		},
		{
			name: "second_half_active",
			args: args{
				validators: []*silapb.Validator{
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
				},
				indices: []primitives.ValidatorIndex{5, 6, 7, 8, 9},
				seed:    seed,
			},
			want: 7,
		},
		{
			name:             "electra_probability_changes",
			isElectraOrAbove: true,
			args: args{
				validators: []*silapb.Validator{
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalanceElectra},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
				},
				indices: []primitives.ValidatorIndex{3},
				seed:    seed,
			},
			want: 3,
		},
		{
			name:             "electra_probability_changes_all_active",
			isElectraOrAbove: true,
			args: args{
				validators: []*silapb.Validator{
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalanceElectra},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalanceElectra},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance}, // skip this one
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalanceElectra},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalanceElectra},
				},
				indices: []primitives.ValidatorIndex{0, 1, 2, 3, 4},
				seed:    seed,
			},
			want: 4,
		},
		{
			name:             "electra_probability_returns_first_validator_with_criteria",
			isElectraOrAbove: true,
			args: args{
				validators: []*silapb.Validator{
					{EffectiveBalance: 1},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalanceElectra},
					{EffectiveBalance: 1},
					{EffectiveBalance: 1},
					{EffectiveBalance: params.BeaconConfig().MaxEffectiveBalanceElectra},
				},
				indices: []primitives.ValidatorIndex{1},
				seed:    seed,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.ClearCache()

			bState := &silapb.BeaconState{Validators: tt.args.validators}
			stTrie, err := state_native.InitializeFromProtoUnsafePhase0(bState)
			require.NoError(t, err)
			if tt.isElectraOrAbove {
				stTrie, err = state_native.InitializeFromProtoUnsafeElectra(&silapb.BeaconStateElectra{Validators: tt.args.validators})
				require.NoError(t, err)
			}

			got, err := helpers.ComputeProposerIndex(stTrie, tt.args.indices, tt.args.seed)
			if tt.wantedErr != "" {
				assert.ErrorContains(t, tt.wantedErr, err)
				return
			}
			assert.NoError(t, err, "received unexpected error")
			assert.Equal(t, tt.want, got, "ComputeProposerIndex()")
		})
	}
}

func TestIsEligibleForActivationQueue(t *testing.T) {
	tests := []struct {
		name         string
		validator    *silapb.Validator
		currentEpoch primitives.Epoch
		want         bool
	}{
		{
			name:         "Eligible",
			validator:    &silapb.Validator{ActivationEligibilityEpoch: params.BeaconConfig().FarFutureEpoch, EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
			currentEpoch: primitives.Epoch(params.BeaconConfig().ElectraForkEpoch - 1),
			want:         true,
		},
		{
			name:         "Incorrect activation eligibility epoch",
			validator:    &silapb.Validator{ActivationEligibilityEpoch: 1, EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance},
			currentEpoch: primitives.Epoch(params.BeaconConfig().ElectraForkEpoch - 1),
			want:         false,
		},
		{
			name:         "Not enough balance",
			validator:    &silapb.Validator{ActivationEligibilityEpoch: params.BeaconConfig().FarFutureEpoch, EffectiveBalance: 1},
			currentEpoch: primitives.Epoch(params.BeaconConfig().ElectraForkEpoch - 1),
			want:         false,
		},
		{
			name:         "More than max effective balance before electra",
			validator:    &silapb.Validator{ActivationEligibilityEpoch: params.BeaconConfig().FarFutureEpoch, EffectiveBalance: params.BeaconConfig().MaxEffectiveBalance + 1},
			currentEpoch: primitives.Epoch(params.BeaconConfig().ElectraForkEpoch - 1),
			want:         false,
		},
		{
			name:         "More than min activation balance after electra",
			validator:    &silapb.Validator{ActivationEligibilityEpoch: params.BeaconConfig().FarFutureEpoch, EffectiveBalance: params.BeaconConfig().MinActivationBalance + 1},
			currentEpoch: primitives.Epoch(params.BeaconConfig().ElectraForkEpoch),
			want:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.ClearCache()

			v, err := state_native.NewValidator(tt.validator)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, helpers.IsEligibleForActivationQueue(v, tt.currentEpoch), "IsEligibleForActivationQueue()")
		})
	}
}

func TestIsIsEligibleForActivation(t *testing.T) {
	tests := []struct {
		name      string
		validator *silapb.Validator
		state     *silapb.BeaconState
		want      bool
	}{
		{"Eligible",
			&silapb.Validator{ActivationEligibilityEpoch: 1, ActivationEpoch: params.BeaconConfig().FarFutureEpoch},
			&silapb.BeaconState{FinalizedCheckpoint: &silapb.Checkpoint{Epoch: 2}},
			true},
		{"Not yet finalized",
			&silapb.Validator{ActivationEligibilityEpoch: 1, ActivationEpoch: params.BeaconConfig().FarFutureEpoch},
			&silapb.BeaconState{FinalizedCheckpoint: &silapb.Checkpoint{Root: make([]byte, 32)}},
			false},
		{"Incorrect activation epoch",
			&silapb.Validator{ActivationEligibilityEpoch: 1},
			&silapb.BeaconState{FinalizedCheckpoint: &silapb.Checkpoint{Epoch: 2}},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.ClearCache()

			s, err := state_native.InitializeFromProtoPhase0(tt.state)
			require.NoError(t, err)
			assert.Equal(t, tt.want, helpers.IsEligibleForActivation(s, tt.validator), "IsEligibleForActivation()")
		})
	}
}

func computeProposerIndexWithValidators(validators []*silapb.Validator, activeIndices []primitives.ValidatorIndex, seed [32]byte) (primitives.ValidatorIndex, error) {
	length := uint64(len(activeIndices))
	if length == 0 {
		return 0, errors.New("empty active indices list")
	}
	hashFunc := hash.CustomSHA256Hasher()

	for i := uint64(0); ; i++ {
		candidateIndex, err := helpers.ComputeShuffledIndex(primitives.ValidatorIndex(i%length), length, seed, true /* shuffle */)
		if err != nil {
			return 0, err
		}
		candidateIndex = activeIndices[candidateIndex]
		if uint64(candidateIndex) >= uint64(len(validators)) {
			return 0, errors.New("active index out of range")
		}
		b := append(seed[:], bytesutil.Bytes8(i/32)...)
		randomByte := hashFunc(b)[i%32]
		v := validators[candidateIndex]
		var effectiveBal uint64
		if v != nil {
			effectiveBal = v.EffectiveBalance
		}
		if effectiveBal*fieldparams.MaxRandomByte >= params.BeaconConfig().MaxEffectiveBalance*uint64(randomByte) {
			return candidateIndex, nil
		}
	}
}

func TestLastActivatedValidatorIndex_OK(t *testing.T) {
	helpers.ClearCache()

	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{})
	require.NoError(t, err)

	validators := make([]*silapb.Validator, 4)
	balances := make([]uint64, len(validators))
	for i := range uint64(4) {
		validators[i] = &silapb.Validator{
			PublicKey:             make([]byte, params.BeaconConfig().BLSPubkeyLength),
			WithdrawalCredentials: make([]byte, 32),
			EffectiveBalance:      32 * 1e9,
			ExitEpoch:             params.BeaconConfig().FarFutureEpoch,
		}
		balances[i] = validators[i].EffectiveBalance
	}
	require.NoError(t, beaconState.SetValidators(validators))
	require.NoError(t, beaconState.SetBalances(balances))

	index, err := helpers.LastActivatedValidatorIndex(t.Context(), beaconState)
	require.NoError(t, err)
	require.Equal(t, index, primitives.ValidatorIndex(3))
}

func TestProposerIndexFromCheckpoint(t *testing.T) {
	helpers.ClearCache()

	e := primitives.Epoch(2)
	r := [32]byte{'a'}
	root := [32]byte{'b'}
	ids := [32]primitives.ValidatorIndex{}
	slot := primitives.Slot(69) // slot 5 in the Epoch
	ids[5] = primitives.ValidatorIndex(19)
	helpers.ProposerIndicesCache().Set(e, r, ids)
	c := &forkchoicetypes.Checkpoint{Root: root, Epoch: e - 1}
	helpers.ProposerIndicesCache().SetCheckpoint(*c, r)
	id, err := helpers.ProposerIndexAtSlotFromCheckpoint(c, slot)
	require.NoError(t, err)
	require.Equal(t, ids[5], id)
}

func TestHasETH1WithdrawalCredentials(t *testing.T) {
	creds := []byte{0xFA, 0xCC}
	v := &silapb.Validator{WithdrawalCredentials: creds}
	roV, err := state_native.NewValidator(v)
	require.NoError(t, err)
	require.Equal(t, false, roV.HasETH1WithdrawalCredentials())
	creds = []byte{params.BeaconConfig().ETH1AddressWithdrawalPrefixByte, 0xCC}
	v = &silapb.Validator{WithdrawalCredentials: creds}
	roV, err = state_native.NewValidator(v)
	require.NoError(t, err)
	require.Equal(t, true, roV.HasETH1WithdrawalCredentials())
	// No Withdrawal cred
}

func TestHasCompoundingWithdrawalCredential(t *testing.T) {
	tests := []struct {
		name      string
		validator *silapb.Validator
		want      bool
	}{
		{"Has compounding withdrawal credential",
			&silapb.Validator{WithdrawalCredentials: bytesutil.PadTo([]byte{params.BeaconConfig().CompoundingWithdrawalPrefixByte}, 32)},
			true},
		{"Does not have compounding withdrawal credential",
			&silapb.Validator{WithdrawalCredentials: bytesutil.PadTo([]byte{0x00}, 32)},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roV, err := state_native.NewValidator(tt.validator)
			require.NoError(t, err)
			assert.Equal(t, tt.want, roV.HasCompoundingWithdrawalCredentials())
		})
	}
}

func TestHasExecutionWithdrawalCredentials(t *testing.T) {
	tests := []struct {
		name      string
		validator *silapb.Validator
		want      bool
	}{
		{"Has compounding withdrawal credential",
			&silapb.Validator{WithdrawalCredentials: bytesutil.PadTo([]byte{params.BeaconConfig().CompoundingWithdrawalPrefixByte}, 32)},
			true},
		{"Has eth1 withdrawal credential",
			&silapb.Validator{WithdrawalCredentials: bytesutil.PadTo([]byte{params.BeaconConfig().ETH1AddressWithdrawalPrefixByte}, 32)},
			true},
		{"Does not have compounding withdrawal credential or eth1 withdrawal credential",
			&silapb.Validator{WithdrawalCredentials: bytesutil.PadTo([]byte{0x00}, 32)},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roV, err := state_native.NewValidator(tt.validator)
			require.NoError(t, err)
			assert.Equal(t, tt.want, roV.HasExecutionWithdrawalCredentials())
		})
	}
}

func TestIsFullyWithdrawableValidator(t *testing.T) {
	tests := []struct {
		name      string
		validator *silapb.Validator
		balance   uint64
		epoch     primitives.Epoch
		fork      int
		want      bool
	}{
		{
			name: "No ETH1 prefix",
			validator: &silapb.Validator{
				WithdrawalCredentials: []byte{0xFA, 0xCC},
				WithdrawableEpoch:     2,
			},
			balance: params.BeaconConfig().MaxEffectiveBalance,
			epoch:   3,
			want:    false,
		},
		{
			name: "Wrong withdrawable epoch",
			validator: &silapb.Validator{
				WithdrawalCredentials: []byte{params.BeaconConfig().ETH1AddressWithdrawalPrefixByte, 0xCC},
				WithdrawableEpoch:     2,
			},
			balance: params.BeaconConfig().MaxEffectiveBalance,
			epoch:   1,
			want:    false,
		},
		{
			name: "No balance",
			validator: &silapb.Validator{
				WithdrawalCredentials: []byte{params.BeaconConfig().ETH1AddressWithdrawalPrefixByte, 0xCC},
				WithdrawableEpoch:     2,
			},
			balance: 0,
			epoch:   3,
			want:    false,
		},
		{
			name: "Fully withdrawable",
			validator: &silapb.Validator{
				WithdrawalCredentials: []byte{params.BeaconConfig().ETH1AddressWithdrawalPrefixByte, 0xCC},
				WithdrawableEpoch:     2,
			},
			balance: params.BeaconConfig().MaxEffectiveBalance,
			epoch:   3,
			want:    true,
		},
		{
			name: "Fully withdrawable compounding validator electra",
			validator: &silapb.Validator{
				WithdrawalCredentials: []byte{params.BeaconConfig().CompoundingWithdrawalPrefixByte, 0xCC},
				WithdrawableEpoch:     2,
			},
			balance: params.BeaconConfig().MaxEffectiveBalance,
			epoch:   params.BeaconConfig().ElectraForkEpoch,
			fork:    version.Electra,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := state_native.NewValidator(tt.validator)
			require.NoError(t, err)
			assert.Equal(t, tt.want, helpers.IsFullyWithdrawableValidator(v, tt.balance, tt.epoch, tt.fork))
		})
	}
}

func TestIsPartiallyWithdrawableValidator(t *testing.T) {
	tests := []struct {
		name      string
		validator *silapb.Validator
		balance   uint64
		epoch     primitives.Epoch
		fork      int
		want      bool
	}{
		{
			name: "No ETH1 prefix",
			validator: &silapb.Validator{
				EffectiveBalance:      params.BeaconConfig().MaxEffectiveBalance,
				WithdrawalCredentials: []byte{0xFA, 0xCC},
			},
			balance: params.BeaconConfig().MaxEffectiveBalance,
			epoch:   3,
			want:    false,
		},
		{
			name: "No balance",
			validator: &silapb.Validator{
				EffectiveBalance:      params.BeaconConfig().MaxEffectiveBalance,
				WithdrawalCredentials: []byte{params.BeaconConfig().ETH1AddressWithdrawalPrefixByte, 0xCC},
			},
			balance: 0,
			epoch:   3,
			want:    false,
		},
		{
			name: "Partially withdrawable",
			validator: &silapb.Validator{
				EffectiveBalance:      params.BeaconConfig().MaxEffectiveBalance,
				WithdrawalCredentials: []byte{params.BeaconConfig().ETH1AddressWithdrawalPrefixByte, 0xCC},
			},
			balance: params.BeaconConfig().MaxEffectiveBalance * 2,
			epoch:   3,
			want:    true,
		},
		{
			name: "Fully withdrawable vanilla validator electra",
			validator: &silapb.Validator{
				EffectiveBalance:      params.BeaconConfig().MinActivationBalance,
				WithdrawalCredentials: []byte{params.BeaconConfig().ETH1AddressWithdrawalPrefixByte, 0xCC},
			},
			balance: params.BeaconConfig().MinActivationBalance * 2,
			epoch:   params.BeaconConfig().ElectraForkEpoch,
			fork:    version.Electra,
			want:    true,
		},
		{
			name: "Fully withdrawable compounding validator electra",
			validator: &silapb.Validator{
				EffectiveBalance:      params.BeaconConfig().MaxEffectiveBalanceElectra,
				WithdrawalCredentials: []byte{params.BeaconConfig().CompoundingWithdrawalPrefixByte, 0xCC},
			},
			balance: params.BeaconConfig().MaxEffectiveBalanceElectra * 2,
			epoch:   params.BeaconConfig().ElectraForkEpoch,
			fork:    version.Electra,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := state_native.NewValidator(tt.validator)
			require.NoError(t, err)
			assert.Equal(t, tt.want, helpers.IsPartiallyWithdrawableValidator(v, tt.balance, tt.epoch, tt.fork))
		})
	}
}

func TestIsSameWithdrawalCredentials(t *testing.T) {
	makeWithdrawalCredentials := func(address []byte) []byte {
		b := make([]byte, 12)
		return append(b, address...)
	}

	tests := []struct {
		name string
		a    *silapb.Validator
		b    *silapb.Validator
		want bool
	}{
		{
			"Same credentials",
			&silapb.Validator{WithdrawalCredentials: makeWithdrawalCredentials([]byte("same"))},
			&silapb.Validator{WithdrawalCredentials: makeWithdrawalCredentials([]byte("same"))},
			true,
		},
		{
			"Different credentials",
			&silapb.Validator{WithdrawalCredentials: makeWithdrawalCredentials([]byte("foo"))},
			&silapb.Validator{WithdrawalCredentials: makeWithdrawalCredentials([]byte("bar"))},
			false,
		},
		{"Handles nil case", nil, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, helpers.IsSameWithdrawalCredentials(tt.a, tt.b))
		})
	}
}

func TestValidatorMaxEffectiveBalance(t *testing.T) {
	tests := []struct {
		name      string
		validator *silapb.Validator
		want      uint64
	}{
		{
			name:      "Compounding withdrawal credential",
			validator: &silapb.Validator{WithdrawalCredentials: []byte{params.BeaconConfig().CompoundingWithdrawalPrefixByte, 0xCC}},
			want:      params.BeaconConfig().MaxEffectiveBalanceElectra,
		},
		{
			name:      "Vanilla credentials",
			validator: &silapb.Validator{WithdrawalCredentials: []byte{params.BeaconConfig().ETH1AddressWithdrawalPrefixByte, 0xCC}},
			want:      params.BeaconConfig().MinActivationBalance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := state_native.NewValidator(tt.validator)
			require.NoError(t, err)
			assert.Equal(t, tt.want, helpers.ValidatorMaxEffectiveBalance(v))
		})
	}
	// Sanity check that MinActivationBalance equals (pre-electra) MaxEffectiveBalance
	assert.Equal(t, params.BeaconConfig().MinActivationBalance, params.BeaconConfig().MaxEffectiveBalance)
}

func TestBeaconProposerIndexAtSlotFulu(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.FuluForkEpoch = 1
	params.OverrideBeaconConfig(cfg)
	lookahead := make([]primitives.ValidatorIndex, 64)
	lookahead[0] = 15
	lookahead[1] = 16
	lookahead[34] = 42
	pbState := silapb.BeaconStateFulu{
		Slot:              100,
		ProposerLookahead: lookahead,
	}
	st, err := state_native.InitializeFromProtoFulu(&pbState)
	require.NoError(t, err)
	idx, err := helpers.BeaconProposerIndexAtSlot(t.Context(), st, 96)
	require.NoError(t, err)
	require.Equal(t, primitives.ValidatorIndex(15), idx)
	idx, err = helpers.BeaconProposerIndexAtSlot(t.Context(), st, 97)
	require.NoError(t, err)
	require.Equal(t, primitives.ValidatorIndex(16), idx)
	idx, err = helpers.BeaconProposerIndexAtSlot(t.Context(), st, 130)
	require.NoError(t, err)
	require.Equal(t, primitives.ValidatorIndex(42), idx)
}
