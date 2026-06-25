package ssz_static

import (
	"context"
	"errors"
	"testing"

	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	// silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/ssz_static"
	fssz "github.com/sila-chain/fastssz"
)

// RunSSZStaticTests executes "ssz_static" tests.
func RunSSZStaticTests(t *testing.T, config string) {
	common.RunSSZStaticTests(t, config, "gloas", unmarshalledSSZ, customHtr)
}

func customHtr(t *testing.T, htrs []common.HTR, object any) []common.HTR {
	_, ok := object.(*silapb.BeaconStateGloas)
	if !ok {
		return htrs
	}

	htrs = append(htrs, func(s any) ([32]byte, error) {
		beaconState, err := state_native.InitializeFromProtoUnsafeGloas(s.(*silapb.BeaconStateGloas))
		require.NoError(t, err)

		return beaconState.HashTreeRoot(context.Background())
	})

	return htrs
}

// unmarshalledSSZ unmarshalls serialized input.
func unmarshalledSSZ(t *testing.T, serializedBytes []byte, folderName string) (any, error) {
	var obj any

	switch folderName {
	// Gloas specific types
	case "SilaPayloadBid":
		obj = &silapb.SilaPayloadBid{}
	case "SignedSilaPayloadBid":
		obj = &silapb.SignedSilaPayloadBid{}
	case "PayloadAttestationData":
		obj = &silapb.PayloadAttestationData{}
	case "PayloadAttestation":
		obj = &silapb.PayloadAttestation{}
	case "PayloadAttestationMessage":
		obj = &silapb.PayloadAttestationMessage{}
	case "BeaconBlock":
		obj = &silapb.BeaconBlockGloas{}
	case "BeaconBlockBody":
		obj = &silapb.BeaconBlockBodyGloas{}
	case "BeaconState":
		obj = &silapb.BeaconStateGloas{}
	case "Builder":
		obj = &silapb.Builder{}
	case "BuilderPendingPayment":
		obj = &silapb.BuilderPendingPayment{}
	case "BuilderPendingWithdrawal":
		obj = &silapb.BuilderPendingWithdrawal{}
	case "SilaPayloadEnvelope":
		obj = &silapb.SilaPayloadEnvelope{}
	case "SignedSilaPayloadEnvelope":
		obj = &silapb.SignedSilaPayloadEnvelope{}
	case "ForkChoiceNode":
		t.Skip("Not a consensus type")
	case "IndexedPayloadAttestation":
		t.Skip("Not a consensus type")
	case "DataColumnSidecar":
		obj = &silapb.DataColumnSidecarGloas{}
	case "SignedProposerPreferences", "ProposerPreferences":
		t.Skip("p2p-only type; not part of the consensus state transition")

	// Standard types that also exist in gloas
	case "SilaPayload":
		obj = &silaenginev1.SilaPayloadGloas{}
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
	case "BeaconBlockHeader":
		obj = &silapb.BeaconBlockHeader{}
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
		obj = &silapb.SignedBeaconBlockGloas{}
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
	case "LightClientOptimisticUpdate", "LightClientFinalityUpdate", "LightClientBootstrap", "LightClientUpdate", "LightClientHeader":
		t.Skip("Gloas light client types not yet implemented")
	case "BlobIdentifier":
		obj = &silapb.BlobIdentifier{}
	case "BlobSidecar":
		t.Skip("Unused type")
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
	case "DataColumnsByRootIdentifier":
		obj = &silapb.DataColumnsByRootIdentifier{}
	case "MatrixEntry":
		t.Skip("Unused type")
	case "PartialDataColumnHeader", "PartialDataColumnPartsMetadata", "PartialDataColumnSidecar", "PartialDataColumnGroupID":
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
