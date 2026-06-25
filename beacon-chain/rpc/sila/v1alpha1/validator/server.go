// Package validator defines a gRPC validator service implementation, providing
// critical endpoints for validator clients to submit blocks/attestations to the
// beacon node, receive assignments, and more.
package validator

import (
	"bytes"
	"context"
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/builder"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache/depositsnapshot"
	blockfeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/block"
	opfeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/operation"
	statefeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/execution"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/attestations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/blstoexec"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/payloadattestation"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/slashings"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/synccommittee"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/voluntaryexits"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/rpc/core"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/startup"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stategen"
	silaSync "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/sync"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/genesis"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server defines a server implementation of the gRPC Validator service,
// providing RPC endpoints for obtaining validator assignments per epoch, the slots
// and committees in which particular validators need to perform their responsibilities,
// and more.
type Server struct {
	Ctx                              context.Context
	PayloadIDCache                   *cache.PayloadIDCache
	TrackedValidatorsCache           *cache.TrackedValidatorsCache
	ProposerPreferencesCache         *cache.ProposerPreferencesCache
	HighestBidCache                  *cache.HighestSilaPayloadBidCache
	SilaPayloadEnvelopeCache    *cache.SilaPayloadEnvelopeCache
	HeadFetcher                      blockchain.HeadFetcher
	ForkFetcher                      blockchain.ForkFetcher
	ForkchoiceFetcher                blockchain.ForkchoiceFetcher
	GenesisFetcher                   blockchain.GenesisFetcher
	FinalizationFetcher              blockchain.FinalizationFetcher
	TimeFetcher                      blockchain.TimeFetcher
	BlockFetcher                     execution.POWBlockFetcher
	DepositFetcher                   cache.DepositFetcher
	ChainStartFetcher                execution.ChainStartFetcher
	SilaExecutionInfoFetcher                  execution.ChainInfoFetcher
	OptimisticModeFetcher            blockchain.OptimisticModeFetcher
	SyncChecker                      silaSync.Checker
	StateNotifier                    statefeed.Notifier
	BlockNotifier                    blockfeed.Notifier
	P2P                              p2p.Broadcaster
	AttestationCache                 *cache.AttestationCache
	AttPool                          attestations.Pool
	PayloadAttestationPool           payloadattestation.PoolManager
	SlashingsPool                    slashings.PoolManager
	ExitPool                         voluntaryexits.PoolManager
	SyncCommitteePool                synccommittee.Pool
	BlockReceiver                    blockchain.BlockReceiver
	PayloadAttestationReceiver       blockchain.PayloadAttestationReceiver
	SilaPayloadEnvelopeReceiver blockchain.SilaPayloadEnvelopeReceiver
	BlobReceiver                     blockchain.BlobReceiver
	DataColumnReceiver               blockchain.DataColumnReceiver
	MockSilaExecutionVotes                    bool
	SilaExecutionBlockFetcher                 execution.POWBlockFetcher
	PendingDepositsFetcher           depositsnapshot.PendingDepositsFetcher
	OperationNotifier                opfeed.Notifier
	StateGen                         stategen.StateManager
	ReplayerBuilder                  stategen.ReplayerBuilder
	BeaconDB                         db.HeadAccessDatabase
	SilaEngineCaller            execution.EngineCaller
	BlockBuilder                     builder.BlockBuilder
	BLSChangesPool                   blstoexec.PoolManager
	ClockWaiter                      startup.ClockWaiter
	CoreService                      *core.Service
	AttestationStateFetcher          blockchain.AttestationStateFetcher
	GraffitiInfo                     *execution.GraffitiInfo
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// WaitForActivation checks if a validator public key exists in the active validator registry of the current
// beacon state, if not, then it creates a stream which listens for canonical states which contain
// the validator with the public key as an active validator record.
// Deprecated: do not use, just poll validator status every epoch.
func (vs *Server) WaitForActivation(req *silapb.ValidatorActivationRequest, stream silapb.BeaconNodeValidator_WaitForActivationServer) error {
	activeValidatorExists, validatorStatuses, err := vs.activationStatus(stream.Context(), req.PublicKeys)
	if err != nil {
		return status.Errorf(codes.Internal, "Could not fetch validator status: %v", err)
	}
	res := &silapb.ValidatorActivationResponse{
		Statuses: validatorStatuses,
	}
	if activeValidatorExists {
		return stream.Send(res)
	}
	if err := stream.Send(res); err != nil {
		return status.Errorf(codes.Internal, "Could not send response over stream: %v", err)
	}

	waitTime := time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second
	ticker := time.NewTicker(waitTime)
	defer ticker.Stop()

	for {
		select {
		// Pinging every slot for activation.
		case <-ticker.C:
			activeValidatorExists, validatorStatuses, err := vs.activationStatus(stream.Context(), req.PublicKeys)
			if err != nil {
				return status.Errorf(codes.Internal, "Could not fetch validator status: %v", err)
			}
			res := &silapb.ValidatorActivationResponse{
				Statuses: validatorStatuses,
			}
			if activeValidatorExists {
				return stream.Send(res)
			}
			if err := stream.Send(res); err != nil {
				return status.Errorf(codes.Internal, "Could not send response over stream: %v", err)
			}
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Stream context canceled")
		case <-vs.Ctx.Done():
			return status.Error(codes.Canceled, "RPC context canceled")
		}
	}
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// ValidatorIndex is called by a validator to get its index location in the beacon state.
func (vs *Server) ValidatorIndex(ctx context.Context, req *silapb.ValidatorIndexRequest) (*silapb.ValidatorIndexResponse, error) {
	st, err := vs.HeadFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not determine head state: %v", err)
	}
	if st == nil || st.IsNil() {
		return nil, status.Errorf(codes.Internal, "head state is empty")
	}
	index, ok := st.ValidatorIndexByPubkey(bytesutil.ToBytes48(req.PublicKey))
	if !ok {
		return nil, status.Errorf(codes.NotFound, "Could not find validator index for public key %#x", req.PublicKey)
	}

	return &silapb.ValidatorIndexResponse{Index: index}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// DomainData fetches the current domain version information from the beacon state.
func (vs *Server) DomainData(ctx context.Context, request *silapb.DomainRequest) (*silapb.DomainResponse, error) {
	epoch := request.Epoch
	rd := bytesutil.ToBytes4(request.Domain)
	if bytes.Equal(request.Domain, params.BeaconConfig().DomainVoluntaryExit[:]) {
		hs, err := vs.HeadFetcher.HeadStateReadOnly(ctx)
		if err != nil {
			return nil, err
		}
		if slots.ToEpoch(hs.Slot()) >= params.BeaconConfig().DenebForkEpoch {
			return computeDomainData(rd, epoch, &silapb.Fork{
				PreviousVersion: params.BeaconConfig().CapellaForkVersion,
				CurrentVersion:  params.BeaconConfig().CapellaForkVersion,
				Epoch:           params.BeaconConfig().CapellaForkEpoch,
			})
		}
	}
	return computeDomainData(rd, epoch, params.ForkFromConfig(params.BeaconConfig(), epoch))
}

func computeDomainData(domain [4]byte, epoch primitives.Epoch, fork *silapb.Fork) (*silapb.DomainResponse, error) {
	gvr := genesis.ValidatorsRoot()
	domainData, err := signing.Domain(fork, epoch, domain, gvr[:])
	if err != nil {
		return nil, err
	}
	return &silapb.DomainResponse{SignatureDomain: domainData}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// WaitForChainStart queries the logs of the Deposit Contract in order to verify the beacon chain
// has started its runtime and validators begin their responsibilities. If it has not, it then
// subscribes to an event stream triggered by the powchain service whenever the ChainStart log does
// occur in the Deposit Contract on ETH 1.0.
func (vs *Server) WaitForChainStart(_ *emptypb.Empty, stream silapb.BeaconNodeValidator_WaitForChainStartServer) error {
	head, err := vs.HeadFetcher.HeadStateReadOnly(stream.Context())
	if err != nil {
		return status.Errorf(codes.Internal, "Could not retrieve head state: %v", err)
	}
	if head != nil && !head.IsNil() {
		res := &silapb.ChainStartResponse{
			Started:               true,
			GenesisTime:           uint64(head.GenesisTime().Unix()),
			GenesisValidatorsRoot: head.GenesisValidatorsRoot(),
		}
		return stream.Send(res)
	}

	clock, err := vs.ClockWaiter.WaitForClock(vs.Ctx)
	if err != nil {
		return status.Error(codes.Canceled, "Context canceled")
	}
	log.WithField("startTime", clock.GenesisTime()).Debug("Received chain started event")
	log.Debug("Sending genesis time notification to connected validator clients")
	gvr := clock.GenesisValidatorsRoot()
	res := &silapb.ChainStartResponse{
		Started:               true,
		GenesisTime:           uint64(clock.GenesisTime().Unix()),
		GenesisValidatorsRoot: gvr[:],
	}
	return stream.Send(res)
}
