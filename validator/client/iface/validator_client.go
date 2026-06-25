package iface

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/client/event"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila/common/hexutil"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

type BeaconCommitteeSelection struct {
	SelectionProof []byte
	Slot           primitives.Slot
	ValidatorIndex primitives.ValidatorIndex
}

type beaconCommitteeSelectionJson struct {
	SelectionProof string `json:"selection_proof"`
	Slot           string `json:"slot"`
	ValidatorIndex string `json:"validator_index"`
}

func (b *BeaconCommitteeSelection) MarshalJSON() ([]byte, error) {
	return json.Marshal(beaconCommitteeSelectionJson{
		SelectionProof: hexutil.Encode(b.SelectionProof),
		Slot:           strconv.FormatUint(uint64(b.Slot), 10),
		ValidatorIndex: strconv.FormatUint(uint64(b.ValidatorIndex), 10),
	})
}

func (b *BeaconCommitteeSelection) UnmarshalJSON(input []byte) error {
	var bjson beaconCommitteeSelectionJson
	err := json.Unmarshal(input, &bjson)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal beacon committee selection")
	}

	slot, err := strconv.ParseUint(bjson.Slot, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse slot")
	}

	vIdx, err := strconv.ParseUint(bjson.ValidatorIndex, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse validator index")
	}

	selectionProof, err := hexutil.Decode(bjson.SelectionProof)
	if err != nil {
		return errors.Wrap(err, "failed to parse selection proof")
	}

	b.Slot = primitives.Slot(slot)
	b.SelectionProof = selectionProof
	b.ValidatorIndex = primitives.ValidatorIndex(vIdx)

	return nil
}

type SyncCommitteeSelection struct {
	SelectionProof    []byte
	Slot              primitives.Slot
	SubcommitteeIndex primitives.CommitteeIndex
	ValidatorIndex    primitives.ValidatorIndex
}

type syncCommitteeSelectionJson struct {
	SelectionProof    string `json:"selection_proof"`
	Slot              string `json:"slot"`
	SubcommitteeIndex string `json:"subcommittee_index"`
	ValidatorIndex    string `json:"validator_index"`
}

func (s *SyncCommitteeSelection) MarshalJSON() ([]byte, error) {
	return json.Marshal(syncCommitteeSelectionJson{
		SelectionProof:    hexutil.Encode(s.SelectionProof),
		Slot:              strconv.FormatUint(uint64(s.Slot), 10),
		SubcommitteeIndex: strconv.FormatUint(uint64(s.SubcommitteeIndex), 10),
		ValidatorIndex:    strconv.FormatUint(uint64(s.ValidatorIndex), 10),
	})
}

func (s *SyncCommitteeSelection) UnmarshalJSON(input []byte) error {
	var resJson syncCommitteeSelectionJson
	err := json.Unmarshal(input, &resJson)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal sync committee selection")
	}

	slot, err := strconv.ParseUint(resJson.Slot, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse slot")
	}

	vIdx, err := strconv.ParseUint(resJson.ValidatorIndex, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse validator index")
	}

	subcommIdx, err := strconv.ParseUint(resJson.SubcommitteeIndex, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse subcommittee index")
	}

	selectionProof, err := hexutil.Decode(resJson.SelectionProof)
	if err != nil {
		return errors.Wrap(err, "failed to parse selection proof")
	}

	s.Slot = primitives.Slot(slot)
	s.SelectionProof = selectionProof
	s.ValidatorIndex = primitives.ValidatorIndex(vIdx)
	s.SubcommitteeIndex = primitives.CommitteeIndex(subcommIdx)

	return nil
}

type ValidatorClient interface {
	// Duties is the pre-GLOAS combined endpoint (GetDuties/GetDutiesV2).
	// This will be eventually replaced by the split duty endpoints below.
	Duties(ctx context.Context, in *silapb.DutiesRequest) (*silapb.ValidatorDutiesContainer, error)
	// Split duty endpoints used post-GLOAS.
	AttesterDuties(ctx context.Context, epoch primitives.Epoch, validatorIndices []primitives.ValidatorIndex) (*silapb.AttesterDutiesResponse, error)
	ProposerDuties(ctx context.Context, epoch primitives.Epoch) (*silapb.ProposerDutiesResponse, error)
	SyncCommitteeDuties(ctx context.Context, epoch primitives.Epoch, validatorIndices []primitives.ValidatorIndex) (*silapb.SyncCommitteeDutiesResponse, error)
	PTCDuties(ctx context.Context, epoch primitives.Epoch, validatorIndices []primitives.ValidatorIndex) (*silapb.PTCDutiesResponse, error)
	DomainData(ctx context.Context, in *silapb.DomainRequest) (*silapb.DomainResponse, error)
	WaitForChainStart(ctx context.Context, in *empty.Empty) (*silapb.ChainStartResponse, error)
	ValidatorIndex(ctx context.Context, in *silapb.ValidatorIndexRequest) (*silapb.ValidatorIndexResponse, error)
	ValidatorStatus(ctx context.Context, in *silapb.ValidatorStatusRequest) (*silapb.ValidatorStatusResponse, error)
	MultipleValidatorStatus(ctx context.Context, in *silapb.MultipleValidatorStatusRequest) (*silapb.MultipleValidatorStatusResponse, error)
	BeaconBlock(ctx context.Context, in *silapb.BlockRequest) (*silapb.GenericBeaconBlock, error)
	ProposeBeaconBlock(ctx context.Context, in *silapb.GenericSignedBeaconBlock) (*silapb.ProposeResponse, error)
	PrepareBeaconProposer(ctx context.Context, in *silapb.PrepareBeaconProposerRequest) (*empty.Empty, error)
	FeeRecipientByPubKey(ctx context.Context, in *silapb.FeeRecipientByPubKeyRequest) (*silapb.FeeRecipientByPubKeyResponse, error)
	AttestationData(ctx context.Context, in *silapb.AttestationDataRequest) (*silapb.AttestationData, error)
	ProposeAttestation(ctx context.Context, in *silapb.Attestation) (*silapb.AttestResponse, error)
	ProposeAttestationElectra(ctx context.Context, in *silapb.SingleAttestation) (*silapb.AttestResponse, error)
	SubmitAggregateSelectionProof(ctx context.Context, in *silapb.AggregateSelectionRequest, index primitives.ValidatorIndex, committeeLength uint64) (*silapb.AggregateSelectionResponse, error)
	SubmitAggregateSelectionProofElectra(ctx context.Context, in *silapb.AggregateSelectionRequest, _ primitives.ValidatorIndex, _ uint64) (*silapb.AggregateSelectionElectraResponse, error)
	SubmitSignedAggregateSelectionProof(ctx context.Context, in *silapb.SignedAggregateSubmitRequest) (*silapb.SignedAggregateSubmitResponse, error)
	SubmitSignedAggregateSelectionProofElectra(ctx context.Context, in *silapb.SignedAggregateSubmitElectraRequest) (*silapb.SignedAggregateSubmitResponse, error)
	ProposeExit(ctx context.Context, in *silapb.SignedVoluntaryExit) (*silapb.ProposeExitResponse, error)
	SubscribeCommitteeSubnets(ctx context.Context, in *silapb.CommitteeSubnetsSubscribeRequest, duties []*silapb.ValidatorDuty) (*empty.Empty, error)
	CheckDoppelGanger(ctx context.Context, in *silapb.DoppelGangerRequest) (*silapb.DoppelGangerResponse, error)
	SyncMessageBlockRoot(ctx context.Context, in *empty.Empty) (*silapb.SyncMessageBlockRootResponse, error)
	SubmitSyncMessage(ctx context.Context, in *silapb.SyncCommitteeMessage) (*empty.Empty, error)
	SyncSubcommitteeIndex(ctx context.Context, in *silapb.SyncSubcommitteeIndexRequest) (*silapb.SyncSubcommitteeIndexResponse, error)
	SyncCommitteeContribution(ctx context.Context, in *silapb.SyncCommitteeContributionRequest) (*silapb.SyncCommitteeContribution, error)
	SubmitSignedContributionAndProof(ctx context.Context, in *silapb.SignedContributionAndProof) (*empty.Empty, error)
	SubmitValidatorRegistrations(ctx context.Context, in *silapb.SignedValidatorRegistrationsV1) (*empty.Empty, error)
	// SubmitSignedProposerPreferences submits proposer preferences for upcoming
	// proposal slots. This replaces PrepareBeaconProposer and SubmitValidatorRegistrations
	// for Gloas+.
	SubmitSignedProposerPreferences(ctx context.Context, in *silapb.SubmitSignedProposerPreferencesRequest) (*empty.Empty, error)
	SubmitSignedExecutionPayloadBid(ctx context.Context, in *silapb.SignedExecutionPayloadBid) (*empty.Empty, error)
	StartEventStream(ctx context.Context, topics []string, eventsChannel chan<- *event.Event)
	EventStreamIsRunning() bool
	AggregatedSelections(ctx context.Context, selections []BeaconCommitteeSelection) ([]BeaconCommitteeSelection, error)
	AggregatedSyncSelections(ctx context.Context, selections []SyncCommitteeSelection) ([]SyncCommitteeSelection, error)
	Host() string
	EnsureReady(ctx context.Context) bool
	GetExecutionPayloadEnvelope(ctx context.Context, slot primitives.Slot, beaconBlockRoot [32]byte) (*silapb.ExecutionPayloadEnvelope, *silapb.WireBlindedExecutionPayloadEnvelope, error)
	PublishExecutionPayloadEnvelope(ctx context.Context, in *silapb.SignedExecutionPayloadEnvelope) (*empty.Empty, error)
	PublishBlindedExecutionPayloadEnvelope(ctx context.Context, in *silapb.SignedWireBlindedExecutionPayloadEnvelope) (*empty.Empty, error)
	PayloadAttestationData(ctx context.Context, slot primitives.Slot) (*silapb.PayloadAttestationData, error)
	SubmitPayloadAttestation(ctx context.Context, in *silapb.PayloadAttestationMessage) (*empty.Empty, error)
}
