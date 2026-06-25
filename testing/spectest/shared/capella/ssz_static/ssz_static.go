package ssz_static

import (
	"context"
	"errors"
	"testing"

	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/ssz_static"
	fssz "github.com/sila-chain/fastssz"
)

// RunSSZStaticTests executes "ssz_static" tests.
func RunSSZStaticTests(t *testing.T, config string) {
	common.RunSSZStaticTests(t, config, "capella", unmarshalledSSZ, customHtr)
}

func customHtr(t *testing.T, htrs []common.HTR, object any) []common.HTR {
	switch object.(type) {
	case *silapb.BeaconStateCapella:
		htrs = append(htrs, func(s any) ([32]byte, error) {
			beaconState, err := state_native.InitializeFromProtoUnsafeCapella(s.(*silapb.BeaconStateCapella))
			require.NoError(t, err)
			return beaconState.HashTreeRoot(context.Background())
		})
	}
	return htrs
}

// unmarshalledSSZ unmarshalls serialized input.
func unmarshalledSSZ(t *testing.T, serializedBytes []byte, folderName string) (any, error) {
	var obj any
	switch folderName {
	case "SilaPayload":
		obj = &silaenginev1.SilaPayloadCapella{}
	case "SilaPayloadHeader":
		obj = &silaenginev1.SilaPayloadHeaderCapella{}
	case "Attestation":
		obj = &silapb.Attestation{}
	case "AttestationData":
		obj = &silapb.AttestationData{}
	case "AttesterSlashing":
		obj = &silapb.AttesterSlashing{}
	case "AggregateAndProof":
		obj = &silapb.AggregateAttestationAndProof{}
	case "BeaconBlock":
		obj = &silapb.BeaconBlockCapella{}
	case "BeaconBlockBody":
		obj = &silapb.BeaconBlockBodyCapella{}
	case "BeaconBlockHeader":
		obj = &silapb.BeaconBlockHeader{}
	case "BeaconState":
		obj = &silapb.BeaconStateCapella{}
	case "Checkpoint":
		obj = &silapb.Checkpoint{}
	case "Deposit":
		obj = &silapb.Deposit{}
	case "DepositMessage":
		obj = &silapb.DepositMessage{}
	case "DepositData":
		obj = &silapb.Deposit_Data{}
	case "SilaExecutionData":
		obj = &silapb.SilaExecutionData{}
	case "SilaExecutionBlock":
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
	case "LightClientHeader":
		obj = &silapb.LightClientHeaderCapella{}
	case "PendingAttestation":
		obj = &silapb.PendingAttestation{}
	case "ProposerSlashing":
		obj = &silapb.ProposerSlashing{}
	case "SignedAggregateAndProof":
		obj = &silapb.SignedAggregateAttestationAndProof{}
	case "SignedBeaconBlock":
		obj = &silapb.SignedBeaconBlockCapella{}
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
	case "SyncCommitteeMessage":
		obj = &silapb.SyncCommitteeMessage{}
	case "SyncCommitteeContribution":
		obj = &silapb.SyncCommitteeContribution{}
	case "ContributionAndProof":
		obj = &silapb.ContributionAndProof{}
	case "SignedContributionAndProof":
		obj = &silapb.SignedContributionAndProof{}
	case "SyncAggregate":
		obj = &silapb.SyncAggregate{}
	case "SyncAggregatorSelectionData":
		obj = &silapb.SyncAggregatorSelectionData{}
	case "SyncCommittee":
		obj = &silapb.SyncCommittee{}
	case "HistoricalSummary":
		obj = &silapb.HistoricalSummary{}
	case "LightClientOptimisticUpdate":
		obj = &silapb.LightClientOptimisticUpdateCapella{}
	case "LightClientFinalityUpdate":
		obj = &silapb.LightClientFinalityUpdateCapella{}
	case "LightClientBootstrap":
		obj = &silapb.LightClientBootstrapCapella{}
	case "LightClientUpdate":
		obj = &silapb.LightClientUpdateCapella{}
	case "PowBlock":
		obj = &silapb.PowBlock{}
	case "Withdrawal":
		obj = &silaenginev1.Withdrawal{}
	case "BLSToExecutionChange":
		obj = &silapb.BLSToExecutionChange{}
	case "SignedBLSToExecutionChange":
		obj = &silapb.SignedBLSToExecutionChange{}
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
