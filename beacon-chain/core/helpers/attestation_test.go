package helpers_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	silaTime "github.com/sila-chain/Sila-Consensus-Core/v7/time"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
)

func TestAttestation_IsAggregator(t *testing.T) {
	t.Run("aggregator", func(t *testing.T) {
		helpers.ClearCache()

		beaconState, privKeys := util.DeterministicGenesisState(t, 100)
		committee, err := helpers.BeaconCommitteeFromState(t.Context(), beaconState, 0, 0)
		require.NoError(t, err)
		sig := privKeys[0].Sign([]byte{'A'})
		agg, err := helpers.IsAggregator(uint64(len(committee)), sig.Marshal())
		require.NoError(t, err)
		assert.Equal(t, true, agg, "Wanted aggregator true")
	})

	t.Run("not aggregator", func(t *testing.T) {
		helpers.ClearCache()

		params.SetupTestConfigCleanup(t)
		params.OverrideBeaconConfig(params.MinimalSpecConfig())
		beaconState, privKeys := util.DeterministicGenesisState(t, 2048)

		committee, err := helpers.BeaconCommitteeFromState(t.Context(), beaconState, 0, 0)
		require.NoError(t, err)
		sig := privKeys[0].Sign([]byte{'A'})
		agg, err := helpers.IsAggregator(uint64(len(committee)), sig.Marshal())
		require.NoError(t, err)
		assert.Equal(t, false, agg, "Wanted aggregator false")
	})
}

func TestAttestation_ComputeSubnetForAttestation(t *testing.T) {
	helpers.ClearCache()

	// Create 10 committees
	committeeCount := uint64(10)
	validatorCount := committeeCount * params.BeaconConfig().TargetCommitteeSize
	validators := make([]*silapb.Validator, validatorCount)

	for i := range validators {
		k := make([]byte, 48)
		copy(k, strconv.Itoa(i))
		validators[i] = &silapb.Validator{
			PublicKey:             k,
			WithdrawalCredentials: make([]byte, 32),
			ExitEpoch:             params.BeaconConfig().FarFutureEpoch,
		}
	}

	state, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		Validators:  validators,
		Slot:        200,
		BlockRoots:  make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot),
		StateRoots:  make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot),
		RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
	})
	require.NoError(t, err)
	valCount, err := helpers.ActiveValidatorCount(t.Context(), state, slots.ToEpoch(34))
	require.NoError(t, err)

	t.Run("Phase 0", func(t *testing.T) {
		att := &silapb.Attestation{
			AggregationBits: []byte{'A'},
			Data: &silapb.AttestationData{
				Slot:            34,
				CommitteeIndex:  4,
				BeaconBlockRoot: []byte{'C'},
			},
			Signature: []byte{'B'},
		}
		sub := helpers.ComputeSubnetForAttestation(valCount, att)
		assert.Equal(t, uint64(6), sub, "Did not get correct subnet for attestation")
	})
	t.Run("Electra", func(t *testing.T) {
		cb := primitives.NewAttestationCommitteeBits()
		cb.SetBitAt(4, true)
		att := &silapb.AttestationElectra{
			AggregationBits: []byte{'A'},
			CommitteeBits:   cb,
			Data: &silapb.AttestationData{
				Slot:            34,
				BeaconBlockRoot: []byte{'C'},
			},
			Signature: []byte{'B'},
		}
		sub := helpers.ComputeSubnetForAttestation(valCount, att)
		assert.Equal(t, uint64(6), sub, "Did not get correct subnet for attestation")
	})
}

func Test_ValidateAttestationTime(t *testing.T) {
	cfg := params.BeaconConfig().Copy()
	cfg.DenebForkEpoch = 5
	params.OverrideBeaconConfig(cfg)
	params.SetupTestConfigCleanup(t)

	if params.BeaconConfig().MaximumGossipClockDisparityDuration() < 200*time.Millisecond {
		t.Fatal("This test expects the maximum clock disparity to be at least 200ms")
	}

	type args struct {
		attSlot     primitives.Slot
		genesisTime time.Time
	}
	tests := []struct {
		name      string
		args      args
		wantedErr string
	}{
		{
			name: "attestation.slot == current_slot",
			args: args{
				attSlot:     15,
				genesisTime: silaTime.Now().Add(-15 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
			},
		},
		{
			name: "attestation.slot == current_slot, received in middle of slot",
			args: args{
				attSlot: 15,
				genesisTime: silaTime.Now().Add(
					-15 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second,
				).Add(-(time.Duration(params.BeaconConfig().SecondsPerSlot/2) * time.Second)),
			},
		},
		{
			name: "attestation.slot == current_slot, received 200ms early",
			args: args{
				attSlot: 16,
				genesisTime: silaTime.Now().Add(
					-16 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second,
				).Add(-200 * time.Millisecond),
			},
		},
		{
			name: "attestation.slot > current_slot",
			args: args{
				attSlot:     16,
				genesisTime: silaTime.Now().Add(-15 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
			},
			wantedErr: "not within attestation propagation range",
		},
		{
			name: "attestation.slot < current_slot-ATTESTATION_PROPAGATION_SLOT_RANGE",
			args: args{
				attSlot:     100 - params.BeaconConfig().AttestationPropagationSlotRange - 1,
				genesisTime: silaTime.Now().Add(-100 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
			},
			wantedErr: "not within attestation propagation range",
		},
		{
			name: "attestation.slot = current_slot-ATTESTATION_PROPAGATION_SLOT_RANGE",
			args: args{
				attSlot:     100 - params.BeaconConfig().AttestationPropagationSlotRange,
				genesisTime: silaTime.Now().Add(-100 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
			},
		},
		{
			name: "attestation.slot = current_slot-ATTESTATION_PROPAGATION_SLOT_RANGE, received 200ms late",
			args: args{
				attSlot: 100 - params.BeaconConfig().AttestationPropagationSlotRange,
				genesisTime: silaTime.Now().Add(
					-100 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second,
				).Add(200 * time.Millisecond),
			},
		},
		{
			name: "attestation.slot < current_slot-ATTESTATION_PROPAGATION_SLOT_RANGE in deneb",
			args: args{
				attSlot:     300 - params.BeaconConfig().AttestationPropagationSlotRange - 1,
				genesisTime: silaTime.Now().Add(-300 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
			},
		},
		{
			name: "attestation.slot = current_slot-ATTESTATION_PROPAGATION_SLOT_RANGE in deneb",
			args: args{
				attSlot:     300 - params.BeaconConfig().AttestationPropagationSlotRange,
				genesisTime: silaTime.Now().Add(-300 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
			},
		},
		{
			name: "attestation.slot = current_slot-ATTESTATION_PROPAGATION_SLOT_RANGE, received 200ms late in deneb",
			args: args{
				attSlot: 300 - params.BeaconConfig().AttestationPropagationSlotRange,
				genesisTime: silaTime.Now().Add(
					-300 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second,
				).Add(200 * time.Millisecond),
			},
		},
		{
			name: "attestation.slot != current epoch or previous epoch in deneb",
			args: args{
				attSlot: 300 - params.BeaconConfig().AttestationPropagationSlotRange,
				genesisTime: silaTime.Now().Add(
					-500 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second,
				).Add(200 * time.Millisecond),
			},
			wantedErr: "attestation epoch 8 not within current epoch 15 or previous epoch",
		},
		{
			name: "attestation.slot is well beyond current slot",
			args: args{
				attSlot:     1024,
				genesisTime: time.Now().Add(-15 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
			},
			wantedErr: "attestation slot 1024 not within attestation propagation range of 0 to 15 (current slot)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.ClearCache()

			err := helpers.ValidateAttestationTime(tt.args.attSlot, tt.args.genesisTime,
				params.BeaconConfig().MaximumGossipClockDisparityDuration())
			if tt.wantedErr != "" {
				assert.ErrorContains(t, tt.wantedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifyCheckpointEpoch_Ok(t *testing.T) {
	helpers.ClearCache()

	// Genesis was 6 epochs ago exactly.
	offset := params.BeaconConfig().SlotsPerEpoch.Mul(params.BeaconConfig().SecondsPerSlot * 6)
	genesis := time.Now().Add(-1 * time.Second * time.Duration(offset))
	assert.Equal(t, true, helpers.VerifyCheckpointEpoch(&silapb.Checkpoint{Epoch: 6}, genesis))
	assert.Equal(t, true, helpers.VerifyCheckpointEpoch(&silapb.Checkpoint{Epoch: 5}, genesis))
	assert.Equal(t, false, helpers.VerifyCheckpointEpoch(&silapb.Checkpoint{Epoch: 4}, genesis))
	assert.Equal(t, false, helpers.VerifyCheckpointEpoch(&silapb.Checkpoint{Epoch: 2}, genesis))
}

func TestValidateNilAttestation(t *testing.T) {
	tests := []struct {
		name        string
		attestation silapb.Att
		errString   string
	}{
		{
			name:        "nil attestation",
			attestation: nil,
			errString:   "attestation is nil",
		},
		{
			name:        "nil attestation data",
			attestation: &silapb.Attestation{},
			errString:   "attestation is nil",
		},
		{
			name: "nil attestation source",
			attestation: &silapb.Attestation{
				Data: &silapb.AttestationData{
					Source: nil,
					Target: &silapb.Checkpoint{},
				},
			},
			errString: "attestation's source can't be nil",
		},
		{
			name: "nil attestation target",
			attestation: &silapb.Attestation{
				Data: &silapb.AttestationData{
					Target: nil,
					Source: &silapb.Checkpoint{},
				},
			},
			errString: "attestation's target can't be nil",
		},
		{
			name: "nil attestation bitfield",
			attestation: &silapb.Attestation{
				Data: &silapb.AttestationData{
					Target: &silapb.Checkpoint{},
					Source: &silapb.Checkpoint{},
				},
			},
			errString: "attestation's bitfield can't be nil",
		},
		{
			name: "good attestation",
			attestation: &silapb.Attestation{
				Data: &silapb.AttestationData{
					Target: &silapb.Checkpoint{},
					Source: &silapb.Checkpoint{},
				},
				AggregationBits: []byte{},
			},
			errString: "",
		},
		{
			name: "single attestation",
			attestation: &silapb.SingleAttestation{
				Data: &silapb.AttestationData{
					Target: &silapb.Checkpoint{},
					Source: &silapb.Checkpoint{},
				},
			},
			errString: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.ClearCache()

			if tt.errString != "" {
				require.ErrorContains(t, tt.errString, helpers.ValidateNilAttestation(tt.attestation))
			} else {
				require.NoError(t, helpers.ValidateNilAttestation(tt.attestation))
			}
		})
	}
}

func TestValidateSlotTargetEpoch(t *testing.T) {
	tests := []struct {
		name        string
		attestation *silapb.Attestation
		errString   string
	}{
		{
			name: "incorrect slot",
			attestation: &silapb.Attestation{
				Data: &silapb.AttestationData{
					Target: &silapb.Checkpoint{Epoch: 1},
					Source: &silapb.Checkpoint{},
				},
				AggregationBits: []byte{},
			},
			errString: "slot 0 does not match target epoch 1",
		},
		{
			name: "good attestation",
			attestation: &silapb.Attestation{
				Data: &silapb.AttestationData{
					Slot:   2 * params.BeaconConfig().SlotsPerEpoch,
					Target: &silapb.Checkpoint{Epoch: 2},
					Source: &silapb.Checkpoint{},
				},
				AggregationBits: []byte{},
			},
			errString: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.ClearCache()

			if tt.errString != "" {
				require.ErrorContains(t, tt.errString, helpers.ValidateSlotTargetEpoch(tt.attestation.Data))
			} else {
				require.NoError(t, helpers.ValidateSlotTargetEpoch(tt.attestation.Data))
			}
		})
	}
}
