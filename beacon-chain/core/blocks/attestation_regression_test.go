package blocks_test

import (
	"context"
	"os"
	"testing"

	"github.com/sila-chain/go-bitfield"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

// Beaconfuzz discovered an off by one issue where an attestation could be produced which would pass
// validation when att.Data.CommitteeIndex is 1 and the committee count per slot is also 1. The only
// valid att.Data.Committee index would be 0, so this is an off by one error.
// See: https://github.com/sigp/beacon-fuzz/issues/78
func TestProcessAttestationNoVerifySignature_BeaconFuzzIssue78(t *testing.T) {
	attData, err := os.ReadFile("testdata/beaconfuzz_78_attestation.ssz")
	if err != nil {
		t.Fatal(err)
	}
	att := &silapb.Attestation{}
	if err := att.UnmarshalSSZ(attData); err != nil {
		t.Fatal(err)
	}
	stateData, err := os.ReadFile("testdata/beaconfuzz_78_beacon.ssz")
	if err != nil {
		t.Fatal(err)
	}
	spb := &silapb.BeaconState{}
	if err := spb.UnmarshalSSZ(stateData); err != nil {
		t.Fatal(err)
	}
	st, err := state_native.InitializeFromProtoUnsafePhase0(spb)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	_, err = blocks.ProcessAttestationNoVerifySignature(ctx, st, att)
	require.ErrorContains(t, "committee index 1 >= committee count 1", err)
}

// Regression introduced in https://github.com/sila-chain/sila/pull/8566.
func TestVerifyAttestationNoVerifySignature_IncorrectSourceEpoch(t *testing.T) {
	// Attestation with an empty signature

	beaconState, _ := util.DeterministicGenesisState(t, 100)

	aggBits := bitfield.NewBitlist(3)
	aggBits.SetBitAt(1, true)
	var mockRoot [32]byte
	copy(mockRoot[:], "hello-world")
	att := &silapb.Attestation{
		Data: &silapb.AttestationData{
			Source: &silapb.Checkpoint{Epoch: 99, Root: mockRoot[:]},
			Target: &silapb.Checkpoint{Epoch: 0, Root: make([]byte, 32)},
		},
		AggregationBits: aggBits,
	}

	var zeroSig [96]byte
	att.Signature = zeroSig[:]

	err := beaconState.SetSlot(beaconState.Slot() + params.BeaconConfig().MinAttestationInclusionDelay)
	require.NoError(t, err)
	ckp := beaconState.CurrentJustifiedCheckpoint()
	copy(ckp.Root, "hello-world")
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(ckp))
	require.NoError(t, beaconState.AppendCurrentEpochAttestations(&silapb.PendingAttestation{}))

	err = blocks.VerifyAttestationNoVerifySignature(context.TODO(), beaconState, att)
	assert.NotEqual(t, nil, err)
}
