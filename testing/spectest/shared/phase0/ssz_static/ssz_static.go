package ssz_static

import (
	"context"
	"errors"
	"testing"

	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/ssz_static"
	fssz "github.com/sila-chain/fastssz"
)

// RunSSZStaticTests executes "ssz_static" tests.
func RunSSZStaticTests(t *testing.T, config string) {
	common.RunSSZStaticTests(t, config, "phase0", unmarshalledSSZ, customHtr)
}

func customHtr(t *testing.T, htrs []common.HTR, object any) []common.HTR {
	switch object.(type) {
	case *silapb.BeaconState:
		htrs = append(htrs, func(s any) ([32]byte, error) {
			beaconState, err := state_native.InitializeFromProtoUnsafePhase0(s.(*silapb.BeaconState))
			require.NoError(t, err)
			return beaconState.HashTreeRoot(context.TODO())
		})
	}
	return htrs
}

// unmarshalledSSZ unmarshalls serialized input.
func unmarshalledSSZ(t *testing.T, serializedBytes []byte, objectName string) (any, error) {
	var obj any
	switch objectName {
	case "Attestation":
		obj = &silapb.Attestation{}
	case "AttestationData":
		obj = &silapb.AttestationData{}
	case "AttesterSlashing":
		obj = &silapb.AttesterSlashing{}
	case "AggregateAndProof":
		obj = &silapb.AggregateAttestationAndProof{}
	case "BeaconBlock":
		obj = &silapb.BeaconBlock{}
	case "BeaconBlockBody":
		obj = &silapb.BeaconBlockBody{}
	case "BeaconBlockHeader":
		obj = &silapb.BeaconBlockHeader{}
	case "BeaconState":
		obj = &silapb.BeaconState{}
	case "Checkpoint":
		obj = &silapb.Checkpoint{}
	case "Deposit":
		obj = &silapb.Deposit{}
	case "DepositMessage":
		obj = &silapb.DepositMessage{}
	case "DepositData":
		obj = &silapb.Deposit_Data{}
	case "Eth1Data":
		obj = &silapb.Eth1Data{}
	case "Eth1Block":
		t.Skip("Unused type")
		return nil, nil
	case "Fork":
		obj = &silapb.Fork{}
	case "ForkData":
		obj = &silapb.ForkData{}
	case "HistoricalBatch":
		obj = &silapb.HistoricalBatch{}
	case "IndexedAttestation":
		obj = &silapb.IndexedAttestation{}
	case "PendingAttestation":
		obj = &silapb.PendingAttestation{}
	case "ProposerSlashing":
		obj = &silapb.ProposerSlashing{}
	case "SignedAggregateAndProof":
		obj = &silapb.SignedAggregateAttestationAndProof{}
	case "SignedBeaconBlock":
		obj = &silapb.SignedBeaconBlock{}
	case "SignedBeaconBlockHeader":
		obj = &silapb.SignedBeaconBlockHeader{}
	case "SignedVoluntaryExit":
		obj = &silapb.SignedVoluntaryExit{}
	case "SigningData":
		obj = &silapb.SigningData{}
	case "Validator":
		obj = &silapb.Validator{}
	case "VoluntaryExit":
		obj = &silapb.VoluntaryExit{}
	default:
		return nil, errors.New("type not found")
	}
	var err error
	if o, ok := obj.(fssz.Unmarshaler); ok {
		err = o.UnmarshalSSZ(serializedBytes)
	} else {
		err = errors.New("could not unmarshal object, not a fastssz compatible object")
	}
	return obj, err
}
