package beacon_api

import (
	"context"
	"net/http"
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/client/event"
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/fallback"
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/rest"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/iface"
	"github.com/sila-chain/Sila/common/hexutil"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

type ValidatorClientOpt func(*beaconApiValidatorClient)

type beaconApiValidatorClient struct {
	genesisProvider         GenesisProvider
	dutiesProvider          dutiesProvider
	stateValidatorsProvider StateValidatorsProvider
	restProvider            rest.RestConnectionProvider
	handler                 rest.Handler
	nodeClient              *beaconApiNodeClient
	beaconBlockConverter    BeaconBlockConverter
	silaChainClient        iface.SilaChainClient
	isEventStreamRunning    bool
	stateless               bool
	envelopeCache           *executionPayloadEnvelopeCache
}

// WithStateless configures the validator client to use the Gloas stateless block production path,
// retrieving the block and execution payload envelope in a single v4 call and caching the envelope
// for reuse by the self-build publisher.
func WithStateless(enabled bool) ValidatorClientOpt {
	return func(c *beaconApiValidatorClient) {
		c.stateless = enabled
		if enabled {
			c.envelopeCache = newExecutionPayloadEnvelopeCache()
		}
	}
}

func NewBeaconApiValidatorClient(provider rest.RestConnectionProvider, opts ...ValidatorClientOpt) iface.ValidatorClient {
	handler := provider.Handler()
	nc := &beaconApiNodeClient{handler: handler}
	c := &beaconApiValidatorClient{
		genesisProvider:         &beaconApiGenesisProvider{handler: handler},
		dutiesProvider:          beaconApiDutiesProvider{handler: handler},
		stateValidatorsProvider: beaconApiStateValidatorsProvider{handler: handler},
		restProvider:            provider,
		handler:                 handler,
		nodeClient:              nc,
		beaconBlockConverter:    beaconApiBeaconBlockConverter{},
		silaChainClient: silaChainClient{
			nodeClient: nc,
			handler:    handler,
		},
		isEventStreamRunning: false,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *beaconApiValidatorClient) Duties(ctx context.Context, in *silapb.DutiesRequest) (*silapb.ValidatorDutiesContainer, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.Duties")
	defer span.End()
	return wrapInMetrics[*silapb.ValidatorDutiesContainer]("Duties", func() (*silapb.ValidatorDutiesContainer, error) {
		return c.duties(ctx, in)
	})
}

func (c *beaconApiValidatorClient) AttesterDuties(ctx context.Context, epoch primitives.Epoch, validatorIndices []primitives.ValidatorIndex) (*silapb.AttesterDutiesResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.AttesterDuties")
	defer span.End()
	return wrapInMetrics[*silapb.AttesterDutiesResponse]("AttesterDuties", func() (*silapb.AttesterDutiesResponse, error) {
		return c.attesterDuties(ctx, epoch, validatorIndices)
	})
}

func (c *beaconApiValidatorClient) ProposerDuties(ctx context.Context, epoch primitives.Epoch) (*silapb.ProposerDutiesResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.ProposerDuties")
	defer span.End()
	return wrapInMetrics[*silapb.ProposerDutiesResponse]("ProposerDuties", func() (*silapb.ProposerDutiesResponse, error) {
		return c.proposerDuties(ctx, epoch)
	})
}

func (c *beaconApiValidatorClient) SyncCommitteeDuties(ctx context.Context, epoch primitives.Epoch, validatorIndices []primitives.ValidatorIndex) (*silapb.SyncCommitteeDutiesResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SyncCommitteeDuties")
	defer span.End()
	return wrapInMetrics[*silapb.SyncCommitteeDutiesResponse]("SyncCommitteeDuties", func() (*silapb.SyncCommitteeDutiesResponse, error) {
		return c.syncCommitteeDuties(ctx, epoch, validatorIndices)
	})
}

func (c *beaconApiValidatorClient) PTCDuties(ctx context.Context, epoch primitives.Epoch, validatorIndices []primitives.ValidatorIndex) (*silapb.PTCDutiesResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.PTCDuties")
	defer span.End()
	return wrapInMetrics[*silapb.PTCDutiesResponse]("PTCDuties", func() (*silapb.PTCDutiesResponse, error) {
		return c.ptcDuties(ctx, epoch, validatorIndices)
	})
}

func (c *beaconApiValidatorClient) CheckDoppelGanger(ctx context.Context, in *silapb.DoppelGangerRequest) (*silapb.DoppelGangerResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.CheckDoppelGanger")
	defer span.End()
	return wrapInMetrics[*silapb.DoppelGangerResponse]("CheckDoppelGanger", func() (*silapb.DoppelGangerResponse, error) {
		return c.checkDoppelGanger(ctx, in)
	})
}

func (c *beaconApiValidatorClient) DomainData(ctx context.Context, in *silapb.DomainRequest) (*silapb.DomainResponse, error) {
	if len(in.Domain) != 4 {
		return nil, errors.Errorf("invalid domain type: %s", hexutil.Encode(in.Domain))
	}

	ctx, span := trace.StartSpan(ctx, "beacon-api.DomainData")
	defer span.End()

	domainType := bytesutil.ToBytes4(in.Domain)

	return wrapInMetrics[*silapb.DomainResponse]("DomainData", func() (*silapb.DomainResponse, error) {
		return c.domainData(ctx, in.Epoch, domainType)
	})
}

func (c *beaconApiValidatorClient) AttestationData(ctx context.Context, in *silapb.AttestationDataRequest) (*silapb.AttestationData, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.AttestationData")
	defer span.End()

	return wrapInMetrics[*silapb.AttestationData]("AttestationData", func() (*silapb.AttestationData, error) {
		return c.attestationData(ctx, in.Slot, in.CommitteeIndex)
	})
}

func (c *beaconApiValidatorClient) BeaconBlock(ctx context.Context, in *silapb.BlockRequest) (*silapb.GenericBeaconBlock, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.BeaconBlock")
	defer span.End()

	return wrapInMetrics[*silapb.GenericBeaconBlock]("BeaconBlock", func() (*silapb.GenericBeaconBlock, error) {
		return c.beaconBlock(ctx, in.Slot, in.RandaoReveal, in.Graffiti)
	})
}

func (c *beaconApiValidatorClient) FeeRecipientByPubKey(_ context.Context, _ *silapb.FeeRecipientByPubKeyRequest) (*silapb.FeeRecipientByPubKeyResponse, error) {
	return nil, nil
}

func (c *beaconApiValidatorClient) SyncCommitteeContribution(ctx context.Context, in *silapb.SyncCommitteeContributionRequest) (*silapb.SyncCommitteeContribution, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SyncCommitteeContribution")
	defer span.End()

	return wrapInMetrics[*silapb.SyncCommitteeContribution]("SyncCommitteeContribution", func() (*silapb.SyncCommitteeContribution, error) {
		return c.syncCommitteeContribution(ctx, in)
	})
}

func (c *beaconApiValidatorClient) SyncMessageBlockRoot(ctx context.Context, _ *empty.Empty) (*silapb.SyncMessageBlockRootResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SyncMessageBlockRoot")
	defer span.End()

	return wrapInMetrics[*silapb.SyncMessageBlockRootResponse]("SyncMessageBlockRoot", func() (*silapb.SyncMessageBlockRootResponse, error) {
		return c.syncMessageBlockRoot(ctx)
	})
}

func (c *beaconApiValidatorClient) SyncSubcommitteeIndex(ctx context.Context, in *silapb.SyncSubcommitteeIndexRequest) (*silapb.SyncSubcommitteeIndexResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SyncSubcommitteeIndex")
	defer span.End()

	return wrapInMetrics[*silapb.SyncSubcommitteeIndexResponse]("SyncSubcommitteeIndex", func() (*silapb.SyncSubcommitteeIndexResponse, error) {
		return c.syncSubcommitteeIndex(ctx, in)
	})
}

func (c *beaconApiValidatorClient) MultipleValidatorStatus(ctx context.Context, in *silapb.MultipleValidatorStatusRequest) (*silapb.MultipleValidatorStatusResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.MultipleValidatorStatus")
	defer span.End()

	return wrapInMetrics[*silapb.MultipleValidatorStatusResponse]("MultipleValidatorStatus", func() (*silapb.MultipleValidatorStatusResponse, error) {
		return c.multipleValidatorStatus(ctx, in)
	})
}

func (c *beaconApiValidatorClient) PrepareBeaconProposer(ctx context.Context, in *silapb.PrepareBeaconProposerRequest) (*empty.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.PrepareBeaconProposer")
	defer span.End()

	return wrapInMetrics[*empty.Empty]("PrepareBeaconProposer", func() (*empty.Empty, error) {
		return new(empty.Empty), c.prepareBeaconProposer(ctx, in.Recipients)
	})
}

func (c *beaconApiValidatorClient) ProposeAttestation(ctx context.Context, in *silapb.Attestation) (*silapb.AttestResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.ProposeAttestation")
	defer span.End()

	return wrapInMetrics[*silapb.AttestResponse]("ProposeAttestation", func() (*silapb.AttestResponse, error) {
		return c.proposeAttestation(ctx, in)
	})
}

func (c *beaconApiValidatorClient) ProposeAttestationElectra(ctx context.Context, in *silapb.SingleAttestation) (*silapb.AttestResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.ProposeAttestationElectra")
	defer span.End()

	return wrapInMetrics[*silapb.AttestResponse]("ProposeAttestationElectra", func() (*silapb.AttestResponse, error) {
		return c.proposeAttestationElectra(ctx, in)
	})
}

func (c *beaconApiValidatorClient) ProposeBeaconBlock(ctx context.Context, in *silapb.GenericSignedBeaconBlock) (*silapb.ProposeResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.ProposeBeaconBlock")
	defer span.End()

	return wrapInMetrics[*silapb.ProposeResponse]("ProposeBeaconBlock", func() (*silapb.ProposeResponse, error) {
		return c.proposeBeaconBlock(ctx, in)
	})
}

func (c *beaconApiValidatorClient) ProposeExit(ctx context.Context, in *silapb.SignedVoluntaryExit) (*silapb.ProposeExitResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.ProposeExit")
	defer span.End()

	return wrapInMetrics[*silapb.ProposeExitResponse]("ProposeExit", func() (*silapb.ProposeExitResponse, error) {
		return c.proposeExit(ctx, in)
	})
}

func (c *beaconApiValidatorClient) StreamBlocksAltair(ctx context.Context, in *silapb.StreamBlocksRequest) (silapb.BeaconNodeValidator_StreamBlocksAltairClient, error) {
	return c.streamBlocks(ctx, in, time.Second), nil
}

func (c *beaconApiValidatorClient) SubmitAggregateSelectionProof(ctx context.Context, in *silapb.AggregateSelectionRequest, index primitives.ValidatorIndex, committeeLength uint64) (*silapb.AggregateSelectionResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SubmitAggregateSelectionProof")
	defer span.End()

	return wrapInMetrics[*silapb.AggregateSelectionResponse]("SubmitAggregateSelectionProof", func() (*silapb.AggregateSelectionResponse, error) {
		return c.submitAggregateSelectionProof(ctx, in, index, committeeLength)
	})
}

func (c *beaconApiValidatorClient) SubmitAggregateSelectionProofElectra(ctx context.Context, in *silapb.AggregateSelectionRequest, index primitives.ValidatorIndex, committeeLength uint64) (*silapb.AggregateSelectionElectraResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SubmitAggregateSelectionProofElectra")
	defer span.End()

	return wrapInMetrics[*silapb.AggregateSelectionElectraResponse]("SubmitAggregateSelectionProofElectra", func() (*silapb.AggregateSelectionElectraResponse, error) {
		return c.submitAggregateSelectionProofElectra(ctx, in, index, committeeLength)
	})
}

func (c *beaconApiValidatorClient) SubmitSignedAggregateSelectionProof(ctx context.Context, in *silapb.SignedAggregateSubmitRequest) (*silapb.SignedAggregateSubmitResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SubmitSignedAggregateSelectionProof")
	defer span.End()

	return wrapInMetrics[*silapb.SignedAggregateSubmitResponse]("SubmitSignedAggregateSelectionProof", func() (*silapb.SignedAggregateSubmitResponse, error) {
		return c.submitSignedAggregateSelectionProof(ctx, in)
	})
}

func (c *beaconApiValidatorClient) SubmitSignedAggregateSelectionProofElectra(ctx context.Context, in *silapb.SignedAggregateSubmitElectraRequest) (*silapb.SignedAggregateSubmitResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SubmitSignedAggregateSelectionProofElectra")
	defer span.End()

	return wrapInMetrics[*silapb.SignedAggregateSubmitResponse]("SubmitSignedAggregateSelectionProofElectra", func() (*silapb.SignedAggregateSubmitResponse, error) {
		return c.submitSignedAggregateSelectionProofElectra(ctx, in)
	})
}

func (c *beaconApiValidatorClient) SubmitSignedContributionAndProof(ctx context.Context, in *silapb.SignedContributionAndProof) (*empty.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SubmitSignedContributionAndProof")
	defer span.End()

	return wrapInMetrics[*empty.Empty]("SubmitSignedContributionAndProof", func() (*empty.Empty, error) {
		return new(empty.Empty), c.submitSignedContributionAndProof(ctx, in)
	})
}

func (c *beaconApiValidatorClient) SubmitSyncMessage(ctx context.Context, in *silapb.SyncCommitteeMessage) (*empty.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SubmitSyncMessage")
	defer span.End()

	return wrapInMetrics[*empty.Empty]("SubmitSyncMessage", func() (*empty.Empty, error) {
		return new(empty.Empty), c.submitSyncMessage(ctx, in)
	})
}

func (c *beaconApiValidatorClient) SubmitValidatorRegistrations(ctx context.Context, in *silapb.SignedValidatorRegistrationsV1) (*empty.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SubmitValidatorRegistrations")
	defer span.End()

	return wrapInMetrics[*empty.Empty]("SubmitValidatorRegistrations", func() (*empty.Empty, error) {
		return new(empty.Empty), c.submitValidatorRegistrations(ctx, in.Messages)
	})
}

func (c *beaconApiValidatorClient) SubmitSignedProposerPreferences(ctx context.Context, in *silapb.SubmitSignedProposerPreferencesRequest) (*empty.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SubmitSignedProposerPreferences")
	defer span.End()

	return wrapInMetrics[*empty.Empty]("SubmitSignedProposerPreferences", func() (*empty.Empty, error) {
		return new(empty.Empty), c.submitSignedProposerPreferences(ctx, in.GetSignedProposerPreferences())
	})
}

// TODO(gloas): Wire up actual REST call to POST /sila/v2/beacon/execution_payload/bid
func (c *beaconApiValidatorClient) SubmitSignedExecutionPayloadBid(_ context.Context, _ *silapb.SignedExecutionPayloadBid) (*empty.Empty, error) {
	log.Debug("SubmitSignedExecutionPayloadBid not yet implemented for beacon API client, skipping")
	return new(empty.Empty), nil
}

func (c *beaconApiValidatorClient) SubscribeCommitteeSubnets(ctx context.Context, in *silapb.CommitteeSubnetsSubscribeRequest, duties []*silapb.ValidatorDuty) (*empty.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SubscribeCommitteeSubnets")
	defer span.End()

	return wrapInMetrics[*empty.Empty]("SubscribeCommitteeSubnets", func() (*empty.Empty, error) {
		return new(empty.Empty), c.subscribeCommitteeSubnets(ctx, in, duties)
	})
}

func (c *beaconApiValidatorClient) ValidatorIndex(ctx context.Context, in *silapb.ValidatorIndexRequest) (*silapb.ValidatorIndexResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.ValidatorIndex")
	defer span.End()

	return wrapInMetrics[*silapb.ValidatorIndexResponse]("ValidatorIndex", func() (*silapb.ValidatorIndexResponse, error) {
		return c.validatorIndex(ctx, in)
	})
}

func (c *beaconApiValidatorClient) ValidatorStatus(ctx context.Context, in *silapb.ValidatorStatusRequest) (*silapb.ValidatorStatusResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.ValidatorStatus")
	defer span.End()

	return c.validatorStatus(ctx, in)
}

// Deprecated: Do not use.
func (c *beaconApiValidatorClient) WaitForChainStart(ctx context.Context, _ *empty.Empty) (*silapb.ChainStartResponse, error) {
	return c.waitForChainStart(ctx)
}

func (c *beaconApiValidatorClient) StartEventStream(ctx context.Context, topics []string, eventsChannel chan<- *event.Event) {
	client := &http.Client{} // event stream should not be subject to the same settings as other api calls
	eventStream, err := event.NewEventStream(ctx, client, c.handler.Host(), topics)
	if err != nil {
		eventsChannel <- &event.Event{
			EventType: event.EventError,
			Data:      []byte(errors.Wrap(err, "failed to start event stream").Error()),
		}
		return
	}
	c.isEventStreamRunning = true
	eventStream.Subscribe(eventsChannel)
	c.isEventStreamRunning = false
}

func (c *beaconApiValidatorClient) EventStreamIsRunning() bool {
	return c.isEventStreamRunning
}

func (c *beaconApiValidatorClient) AggregatedSelections(ctx context.Context, selections []iface.BeaconCommitteeSelection) ([]iface.BeaconCommitteeSelection, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.AggregatedSelections")
	defer span.End()

	return wrapInMetrics[[]iface.BeaconCommitteeSelection]("AggregatedSelections", func() ([]iface.BeaconCommitteeSelection, error) {
		return c.aggregatedSelection(ctx, selections)
	})
}

func (c *beaconApiValidatorClient) AggregatedSyncSelections(ctx context.Context, selections []iface.SyncCommitteeSelection) ([]iface.SyncCommitteeSelection, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.AggregatedSyncSelections")
	defer span.End()

	return wrapInMetrics[[]iface.SyncCommitteeSelection]("AggregatedSyncSelections", func() ([]iface.SyncCommitteeSelection, error) {
		return c.aggregatedSyncSelections(ctx, selections)
	})
}

func wrapInMetrics[Resp any](action string, f func() (Resp, error)) (Resp, error) {
	now := time.Now()
	resp, err := f()
	recordMetrics(action, now, err)
	return resp, err
}

func wrapInMetrics2[R1, R2 any](action string, f func() (R1, R2, error)) (R1, R2, error) {
	now := time.Now()
	r1, r2, err := f()
	recordMetrics(action, now, err)
	return r1, r2, err
}

func recordMetrics(action string, start time.Time, err error) {
	httpActionCount.WithLabelValues(action).Inc()
	if err == nil {
		httpActionLatency.WithLabelValues(action).Observe(time.Since(start).Seconds())
	} else {
		failedHTTPActionCount.WithLabelValues(action).Inc()
	}
}

func (c *beaconApiValidatorClient) Host() string {
	return c.handler.Host()
}

func (c *beaconApiValidatorClient) EnsureReady(ctx context.Context) bool {
	return fallback.EnsureReady(ctx, c.restProvider, c.nodeClient)
}

// Gloas Fork Methods

func (c *beaconApiValidatorClient) GetExecutionPayloadEnvelope(ctx context.Context, slot primitives.Slot, beaconBlockRoot [32]byte) (*silapb.ExecutionPayloadEnvelope, *silapb.WireBlindedExecutionPayloadEnvelope, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.GetExecutionPayloadEnvelope")
	defer span.End()

	return wrapInMetrics2("GetExecutionPayloadEnvelope", func() (*silapb.ExecutionPayloadEnvelope, *silapb.WireBlindedExecutionPayloadEnvelope, error) {
		return c.getExecutionPayloadEnvelope(ctx, slot, beaconBlockRoot)
	})
}

func (c *beaconApiValidatorClient) PublishExecutionPayloadEnvelope(ctx context.Context, in *silapb.SignedExecutionPayloadEnvelope) (*empty.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.PublishExecutionPayloadEnvelope")
	defer span.End()

	return wrapInMetrics[*empty.Empty]("PublishExecutionPayloadEnvelope", func() (*empty.Empty, error) {
		return c.publishExecutionPayloadEnvelope(ctx, in)
	})
}

func (c *beaconApiValidatorClient) PublishBlindedExecutionPayloadEnvelope(ctx context.Context, in *silapb.SignedWireBlindedExecutionPayloadEnvelope) (*empty.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.PublishBlindedExecutionPayloadEnvelope")
	defer span.End()

	return wrapInMetrics[*empty.Empty]("PublishBlindedExecutionPayloadEnvelope", func() (*empty.Empty, error) {
		return c.publishBlindedExecutionPayloadEnvelope(ctx, in)
	})
}

func (c *beaconApiValidatorClient) PayloadAttestationData(ctx context.Context, slot primitives.Slot) (*silapb.PayloadAttestationData, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.PayloadAttestationData")
	defer span.End()

	return wrapInMetrics[*silapb.PayloadAttestationData]("PayloadAttestationData", func() (*silapb.PayloadAttestationData, error) {
		return c.payloadAttestationData(ctx, slot)
	})
}

func (c *beaconApiValidatorClient) SubmitPayloadAttestation(ctx context.Context, msg *silapb.PayloadAttestationMessage) (*empty.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon-api.SubmitPayloadAttestation")
	defer span.End()

	return wrapInMetrics[*empty.Empty]("SubmitPayloadAttestation", func() (*empty.Empty, error) {
		return new(empty.Empty), c.submitPayloadAttestation(ctx, msg)
	})
}
