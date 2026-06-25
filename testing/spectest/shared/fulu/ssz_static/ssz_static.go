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
	common.RunSSZStaticTests(t, config, "fulu", UnmarshalledSSZ, customHtr)
}

func customHtr(t *testing.T, htrs []common.HTR, object any) []common.HTR {
	_, ok := object.(*silapb.BeaconStateFulu)
	if !ok {
		return htrs
	}

	htrs = append(htrs, func(s any) ([32]byte, error) {
		beaconState, err := state_native.InitializeFromProtoUnsafeFulu(s.(*silapb.BeaconStateFulu))
		require.NoError(t, err)
		return beaconState.HashTreeRoot(context.Background())
	})
	return htrs
}

// UnmarshalledSSZ unmarshalls serialized input.
func UnmarshalledSSZ(t *testing.T, serializedBytes []byte, folderName string) (any, error) {
	var obj any
	switch folderName {
	case "SilaPayload":
		obj = &silaenginev1.SilaPayloadDeneb{}
	case "SilaPayloadHeader":
		obj = &silaenginev1.SilaPayloadHeaderDeneb{}
	case "Attestation":
		obj = &silapb.AttestationElectra{}
	case "AttestationData":
		obj = &silapb.AttestationData{}
	case "AttesterSlashing":
		obj = &silapb.AttesterSlashingElectra{}
	case "AggregateAndProof":
		obj = &silapb.AggregateAttestationAndProofElectra{}
	case "BeaconBlock":
		obj = &silapb.BeaconBlockElectra{}
	case "BeaconBlockBody":
		obj = &silapb.BeaconBlockBodyElectra{}
	case "BeaconBlockHeader":
		obj = &silapb.BeaconBlockHeader{}
	case "BeaconState":
		obj = &silapb.BeaconStateFulu{}
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
	case "SilaBlock":
		t.Skip("Unused type")
		return nil, nil
	case "Fork":
		obj = &silapb.Fork{}
	case "ForkData":
		obj = &silapb.ForkData{}
	case "HistoricalBatch":
		obj = &silapb.HistoricalBatch{}
	case "IndexedAttestation":
		obj = &silapb.IndexedAttestationElectra{}
	case "PendingAttestation":
		obj = &silapb.PendingAttestation{}
	case "ProposerSlashing":
		obj = &silapb.ProposerSlashing{}
	case "SignedAggregateAndProof":
		obj = &silapb.SignedAggregateAttestationAndProofElectra{}
	case "SignedBeaconBlock":
		obj = &silapb.SignedBeaconBlockElectra{}
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
	case "SingleAttestation":
		obj = &silapb.SingleAttestation{}
	case "SyncAggregate":
		obj = &silapb.SyncAggregate{}
	case "SyncAggregatorSelectionData":
		obj = &silapb.SyncAggregatorSelectionData{}
	case "SyncCommittee":
		obj = &silapb.SyncCommittee{}
	case "LightClientOptimisticUpdate":
		obj = &silapb.LightClientOptimisticUpdateDeneb{}
	case "LightClientFinalityUpdate":
		obj = &silapb.LightClientFinalityUpdateElectra{}
	case "LightClientBootstrap":
		obj = &silapb.LightClientBootstrapElectra{}
	case "LightClientUpdate":
		obj = &silapb.LightClientUpdateElectra{}
	case "LightClientHeader":
		obj = &silapb.LightClientHeaderDeneb{}
	case "BlobIdentifier":
		obj = &silapb.BlobIdentifier{}
	case "BlobSidecar":
		obj = &silapb.BlobSidecar{}
	case "PowBlock":
		obj = &silapb.PowBlock{}
	case "Withdrawal":
		obj = &silaenginev1.Withdrawal{}
	case "HistoricalSummary":
		obj = &silapb.HistoricalSummary{}
	case "BLSToSilaChange":
		obj = &silapb.BLSToSilaChange{}
	case "SignedBLSToSilaChange":
		obj = &silapb.SignedBLSToSilaChange{}
	case "PendingDeposit":
		obj = &silapb.PendingDeposit{}
	case "PendingPartialWithdrawal":
		obj = &silapb.PendingPartialWithdrawal{}
	case "PendingConsolidation":
		obj = &silapb.PendingConsolidation{}
	case "WithdrawalRequest":
		obj = &silaenginev1.WithdrawalRequest{}
	case "DepositRequest":
		obj = &silaenginev1.DepositRequest{}
	case "ConsolidationRequest":
		obj = &silaenginev1.ConsolidationRequest{}
	case "SilaRequests":
		obj = &silaenginev1.SilaRequests{}
	case "DataColumnSidecar":
		obj = &silapb.DataColumnSidecar{}
	case "DataColumnsByRootIdentifier":
		obj = &silapb.DataColumnsByRootIdentifier{}
	case "MatrixEntry":
		t.Skip("Unused type")
	case "PartialDataColumnHeader", "PartialDataColumnPartsMetadata", "PartialDataColumnSidecar":
		t.Skip("Not yet implemented")
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
