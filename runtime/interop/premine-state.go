package interop

import (
	"context"
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/altair"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stateutil"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/container/trie"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila/core/types"
	"github.com/pkg/errors"
)

var errUnsupportedVersion = errors.New("schema version not supported by PremineGenesisConfig")

type PremineGenesisConfig struct {
	GenesisTime     time.Time
	NVals           uint64
	PregenesisCreds uint64
	Version         int          // as in "github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	GB              *types.Block // geth genesis block
	depositEntries  *depositEntries
}

type depositEntries struct {
	dds   []*silapb.Deposit_Data
	roots [][]byte
}

type PremineGenesisOpt func(*PremineGenesisConfig)

func WithDepositData(dds []*silapb.Deposit_Data, roots [][]byte) PremineGenesisOpt {
	return func(cfg *PremineGenesisConfig) {
		cfg.depositEntries = &depositEntries{
			dds:   dds,
			roots: roots,
		}
	}
}

// NewPreminedGenesis creates a genesis BeaconState at the given fork version, suitable for using as an e2e genesis.
func NewPreminedGenesis(ctx context.Context, genesis time.Time, nvals, pCreds uint64, version int, gb *types.Block, opts ...PremineGenesisOpt) (state.BeaconState, error) {
	cfg := &PremineGenesisConfig{
		GenesisTime:     genesis,
		NVals:           nvals,
		PregenesisCreds: pCreds,
		Version:         version,
		GB:              gb,
	}
	for _, o := range opts {
		o(cfg)
	}
	return cfg.prepare(ctx)
}

func (s *PremineGenesisConfig) prepare(ctx context.Context) (state.BeaconState, error) {
	switch s.Version {
	case version.Phase0, version.Altair, version.Bellatrix, version.Capella, version.Deneb, version.Electra, version.Fulu:
	default:
		return nil, errors.Wrapf(errUnsupportedVersion, "version=%s", version.String(s.Version))
	}

	st, err := s.empty()
	if err != nil {
		return nil, err
	}
	if err = s.processDeposits(ctx, st); err != nil {
		return nil, err
	}
	if err = s.populate(st); err != nil {
		return nil, err
	}

	return st, nil
}

func (s *PremineGenesisConfig) empty() (state.BeaconState, error) {
	var e state.BeaconState
	var err error

	bRoots := make([][]byte, fieldparams.BlockRootsLength)
	for i := range bRoots {
		bRoots[i] = bytesutil.PadTo([]byte{}, 32)
	}
	sRoots := make([][]byte, fieldparams.StateRootsLength)
	for i := range sRoots {
		sRoots[i] = bytesutil.PadTo([]byte{}, 32)
	}
	mixes := make([][]byte, fieldparams.RandaoMixesLength)
	for i := range mixes {
		mixes[i] = bytesutil.PadTo([]byte{}, 32)
	}

	switch s.Version {
	case version.Phase0:
		e, err = state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{
			BlockRoots:  bRoots,
			StateRoots:  sRoots,
			RandaoMixes: mixes,
			Balances:    []uint64{},
			Validators:  []*silapb.Validator{},
		})
		if err != nil {
			return nil, err
		}
	case version.Altair:
		e, err = state_native.InitializeFromProtoUnsafeAltair(&silapb.BeaconStateAltair{
			BlockRoots:       bRoots,
			StateRoots:       sRoots,
			RandaoMixes:      mixes,
			Balances:         []uint64{},
			InactivityScores: []uint64{},
			Validators:       []*silapb.Validator{},
		})
		if err != nil {
			return nil, err
		}
	case version.Bellatrix:
		e, err = state_native.InitializeFromProtoUnsafeBellatrix(&silapb.BeaconStateBellatrix{
			BlockRoots:       bRoots,
			StateRoots:       sRoots,
			RandaoMixes:      mixes,
			Balances:         []uint64{},
			InactivityScores: []uint64{},
			Validators:       []*silapb.Validator{},
		})
		if err != nil {
			return nil, err
		}
	case version.Capella:
		e, err = state_native.InitializeFromProtoUnsafeCapella(&silapb.BeaconStateCapella{
			BlockRoots:       bRoots,
			StateRoots:       sRoots,
			RandaoMixes:      mixes,
			Balances:         []uint64{},
			InactivityScores: []uint64{},
			Validators:       []*silapb.Validator{},
		})
		if err != nil {
			return nil, err
		}
	case version.Deneb:
		e, err = state_native.InitializeFromProtoUnsafeDeneb(&silapb.BeaconStateDeneb{})
		if err != nil {
			return nil, err
		}
	case version.Electra:
		e, err = state_native.InitializeFromProtoUnsafeElectra(&silapb.BeaconStateElectra{
			DepositRequestsStartIndex: params.BeaconConfig().UnsetDepositRequestsStartIndex,
		})
		if err != nil {
			return nil, err
		}
	case version.Fulu:
		e, err = state_native.InitializeFromProtoUnsafeFulu(&silapb.BeaconStateFulu{
			DepositRequestsStartIndex: params.BeaconConfig().UnsetDepositRequestsStartIndex,
		})
		if err != nil {
			return nil, err
		}
	default:
		return nil, errUnsupportedVersion
	}
	if err = e.SetSlot(0); err != nil {
		return nil, err
	}
	if err = e.SetJustificationBits([]byte{0}); err != nil {
		return nil, err
	}
	if err = e.SetHistoricalRoots([][]byte{}); err != nil {
		return nil, err
	}
	zcp := &silapb.Checkpoint{
		Epoch: 0,
		Root:  params.BeaconConfig().ZeroHash[:],
	}
	if err = e.SetPreviousJustifiedCheckpoint(zcp); err != nil {
		return nil, err
	}
	if err = e.SetCurrentJustifiedCheckpoint(zcp); err != nil {
		return nil, err
	}
	if err = e.SetFinalizedCheckpoint(zcp); err != nil {
		return nil, err
	}
	if err = e.SetSilaDataVotes([]*silapb.SilaData{}); err != nil {
		return nil, err
	}
	if s.Version == version.Phase0 {
		if err = e.SetCurrentEpochAttestations([]*silapb.PendingAttestation{}); err != nil {
			return nil, err
		}
		if err = e.SetPreviousEpochAttestations([]*silapb.PendingAttestation{}); err != nil {
			return nil, err
		}
	}
	return e.Copy(), nil
}

func (s *PremineGenesisConfig) processDeposits(ctx context.Context, g state.BeaconState) error {
	deposits, err := s.deposits()
	if err != nil {
		return err
	}
	if err = s.setSilaData(g); err != nil {
		return err
	}
	if _, err = helpers.UpdateGenesisSilaData(g, deposits, g.SilaData()); err != nil {
		return err
	}

	// TODO: should be updated when electra E2E is updated
	_, err = altair.ProcessPreGenesisDeposits(ctx, g, deposits)
	if err != nil {
		return errors.Wrap(err, "could not process validator deposits")
	}
	return nil
}

func (s *PremineGenesisConfig) deposits() ([]*silapb.Deposit, error) {
	if s.depositEntries == nil {
		prv, pub, err := s.keys()
		if err != nil {
			return nil, err
		}
		dds, roots, err := DepositDataFromKeysWithExecCreds(prv, pub, s.PregenesisCreds)
		if err != nil {
			return nil, errors.Wrap(err, "could not generate deposit data from keys")
		}
		s.depositEntries = &depositEntries{
			dds:   dds,
			roots: roots,
		}
	}

	t, err := trie.GenerateTrieFromItems(s.depositEntries.roots, params.BeaconConfig().SilaDepositTreeDepth)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate Merkle trie for deposit proofs")
	}
	deposits, err := GenerateDepositsFromData(s.depositEntries.dds, t)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate deposits from the deposit data provided")
	}
	return deposits, nil
}

func (s *PremineGenesisConfig) keys() ([]bls.SecretKey, []bls.PublicKey, error) {
	prv, pub, err := DeterministicallyGenerateKeys(0, s.NVals)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not deterministically generate keys for %d validators", s.NVals)
	}
	return prv, pub, nil
}

func (s *PremineGenesisConfig) setSilaData(g state.BeaconState) error {
	if err := g.SetSilaExecutionDepositIndex(0); err != nil {
		return err
	}
	dr, err := emptyDepositRoot()
	if err != nil {
		return err
	}
	return g.SetSilaData(&silapb.SilaData{DepositRoot: dr[:], BlockHash: s.GB.Hash().Bytes()})
}

func emptyDepositRoot() ([32]byte, error) {
	t, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	if err != nil {
		return [32]byte{}, err
	}
	return t.HashTreeRoot()
}

func (s *PremineGenesisConfig) populate(g state.BeaconState) error {
	if err := g.SetGenesisTime(s.GenesisTime); err != nil {
		return err
	}
	if err := s.setGenesisValidatorsRoot(g); err != nil {
		return err
	}
	if err := s.setFork(g); err != nil {
		return err
	}
	rao := nSetRoots(uint64(params.BeaconConfig().EpochsPerHistoricalVector), s.GB.Hash().Bytes())
	if err := g.SetRandaoMixes(rao); err != nil {
		return err
	}
	if err := g.SetBlockRoots(nZeroRoots(uint64(params.BeaconConfig().SlotsPerHistoricalRoot))); err != nil {
		return err
	}
	if err := g.SetStateRoots(nZeroRoots(uint64(params.BeaconConfig().SlotsPerHistoricalRoot))); err != nil {
		return err
	}
	if err := g.SetSlashings(make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector)); err != nil {
		return err
	}
	if err := s.setLatestBlockHeader(g); err != nil {
		return err
	}
	if err := s.setInactivityScores(g); err != nil {
		return err
	}
	if err := s.setCurrentEpochParticipation(g); err != nil {
		return err
	}
	if err := s.setPrevEpochParticipation(g); err != nil {
		return err
	}
	if err := s.setSyncCommittees(g); err != nil {
		return err
	}
	if err := s.setSilaPayload(g); err != nil {
		return err
	}

	// For pre-mined genesis, we want to keep the deposit root set to the root of an empty trie.
	// This needs to be set again because the methods used by processDeposits mutate the state's silaData.
	return s.setSilaData(g)
}

func (s *PremineGenesisConfig) setGenesisValidatorsRoot(g state.BeaconState) error {
	compactValidators := stateutil.CompactValidatorsFromProto(g.Validators())
	vroot, err := stateutil.ValidatorRegistryRoot(compactValidators)
	if err != nil {
		return err
	}
	return g.SetGenesisValidatorsRoot(vroot[:])
}

func (s *PremineGenesisConfig) setFork(g state.BeaconState) error {
	var pv, cv []byte
	switch s.Version {
	case version.Phase0:
		pv, cv = params.BeaconConfig().GenesisForkVersion, params.BeaconConfig().GenesisForkVersion
	case version.Altair:
		pv, cv = params.BeaconConfig().GenesisForkVersion, params.BeaconConfig().AltairForkVersion
	case version.Bellatrix:
		pv, cv = params.BeaconConfig().AltairForkVersion, params.BeaconConfig().BellatrixForkVersion
	case version.Capella:
		pv, cv = params.BeaconConfig().BellatrixForkVersion, params.BeaconConfig().CapellaForkVersion
	case version.Deneb:
		pv, cv = params.BeaconConfig().CapellaForkVersion, params.BeaconConfig().DenebForkVersion
	case version.Electra:
		pv, cv = params.BeaconConfig().DenebForkVersion, params.BeaconConfig().ElectraForkVersion
	case version.Fulu:
		pv, cv = params.BeaconConfig().ElectraForkVersion, params.BeaconConfig().FuluForkVersion
	default:
		return errUnsupportedVersion
	}
	fork := &silapb.Fork{
		PreviousVersion: pv,
		CurrentVersion:  cv,
		Epoch:           0,
	}
	return g.SetFork(fork)
}

func (s *PremineGenesisConfig) setInactivityScores(g state.BeaconState) error {
	if s.Version < version.Altair {
		return nil
	}

	scores, err := g.InactivityScores()
	if err != nil {
		return err
	}
	scoresMissing := len(g.Validators()) - len(scores)
	if scoresMissing > 0 {
		for range scoresMissing {
			scores = append(scores, 0)
		}
	}
	return g.SetInactivityScores(scores)
}

func (s *PremineGenesisConfig) setCurrentEpochParticipation(g state.BeaconState) error {
	if s.Version < version.Altair {
		return nil
	}

	p, err := g.CurrentEpochParticipation()
	if err != nil {
		return err
	}
	missing := len(g.Validators()) - len(p)
	if missing > 0 {
		for range missing {
			p = append(p, 0)
		}
	}
	return g.SetCurrentParticipationBits(p)
}

func (s *PremineGenesisConfig) setPrevEpochParticipation(g state.BeaconState) error {
	if s.Version < version.Altair {
		return nil
	}

	p, err := g.PreviousEpochParticipation()
	if err != nil {
		return err
	}
	missing := len(g.Validators()) - len(p)
	if missing > 0 {
		for range missing {
			p = append(p, 0)
		}
	}
	return g.SetPreviousParticipationBits(p)
}

func (s *PremineGenesisConfig) setSyncCommittees(g state.BeaconState) error {
	if s.Version < version.Altair {
		return nil
	}
	sc, err := altair.NextSyncCommittee(context.Background(), g)
	if err != nil {
		return err
	}
	if err = g.SetNextSyncCommittee(sc); err != nil {
		return err
	}
	return g.SetCurrentSyncCommittee(sc)
}

type rooter interface {
	HashTreeRoot() ([32]byte, error)
}

func (s *PremineGenesisConfig) setLatestBlockHeader(g state.BeaconState) error {
	var body rooter
	switch s.Version {
	case version.Phase0:
		body = &silapb.BeaconBlockBody{
			RandaoReveal: make([]byte, 96),
			SilaData: &silapb.SilaData{
				DepositRoot: make([]byte, 32),
				BlockHash:   make([]byte, 32),
			},
			Graffiti: make([]byte, 32),
		}
	case version.Altair:
		body = &silapb.BeaconBlockBodyAltair{
			RandaoReveal: make([]byte, 96),
			SilaData: &silapb.SilaData{
				DepositRoot: make([]byte, 32),
				BlockHash:   make([]byte, 32),
			},
			Graffiti: make([]byte, 32),
			SyncAggregate: &silapb.SyncAggregate{
				SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
				SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
			},
		}
	case version.Bellatrix:
		body = &silapb.BeaconBlockBodyBellatrix{
			RandaoReveal: make([]byte, 96),
			SilaData: &silapb.SilaData{
				DepositRoot: make([]byte, 32),
				BlockHash:   make([]byte, 32),
			},
			Graffiti: make([]byte, 32),
			SyncAggregate: &silapb.SyncAggregate{
				SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
				SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
			},
			SilaPayload: &silaenginev1.SilaPayload{
				ParentHash:    make([]byte, 32),
				FeeRecipient:  make([]byte, 20),
				StateRoot:     make([]byte, 32),
				ReceiptsRoot:  make([]byte, 32),
				LogsBloom:     make([]byte, 256),
				PrevRandao:    make([]byte, 32),
				ExtraData:     make([]byte, 0),
				BaseFeePerGas: make([]byte, 32),
				BlockHash:     make([]byte, 32),
				Transactions:  make([][]byte, 0),
			},
		}
	case version.Capella:
		body = &silapb.BeaconBlockBodyCapella{
			RandaoReveal: make([]byte, 96),
			SilaData: &silapb.SilaData{
				DepositRoot: make([]byte, 32),
				BlockHash:   make([]byte, 32),
			},
			Graffiti: make([]byte, 32),
			SyncAggregate: &silapb.SyncAggregate{
				SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
				SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
			},
			SilaPayload: &silaenginev1.SilaPayloadCapella{
				ParentHash:    make([]byte, 32),
				FeeRecipient:  make([]byte, 20),
				StateRoot:     make([]byte, 32),
				ReceiptsRoot:  make([]byte, 32),
				LogsBloom:     make([]byte, 256),
				PrevRandao:    make([]byte, 32),
				ExtraData:     make([]byte, 0),
				BaseFeePerGas: make([]byte, 32),
				BlockHash:     make([]byte, 32),
				Transactions:  make([][]byte, 0),
				Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
			},
			BlsToSilaChanges: make([]*silapb.SignedBLSToSilaChange, 0),
		}
	case version.Deneb:
		body = &silapb.BeaconBlockBodyDeneb{
			RandaoReveal: make([]byte, 96),
			SilaData: &silapb.SilaData{
				DepositRoot: make([]byte, 32),
				BlockHash:   make([]byte, 32),
			},
			Graffiti: make([]byte, 32),
			SyncAggregate: &silapb.SyncAggregate{
				SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
				SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
			},
			SilaPayload: &silaenginev1.SilaPayloadDeneb{
				ParentHash:    make([]byte, 32),
				FeeRecipient:  make([]byte, 20),
				StateRoot:     make([]byte, 32),
				ReceiptsRoot:  make([]byte, 32),
				LogsBloom:     make([]byte, 256),
				PrevRandao:    make([]byte, 32),
				ExtraData:     make([]byte, 0),
				BaseFeePerGas: make([]byte, 32),
				BlockHash:     make([]byte, 32),
				Transactions:  make([][]byte, 0),
				Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
			},
			BlsToSilaChanges: make([]*silapb.SignedBLSToSilaChange, 0),
			BlobKzgCommitments:    make([][]byte, 0),
		}
	case version.Electra:
		body = &silapb.BeaconBlockBodyElectra{
			RandaoReveal: make([]byte, 96),
			SilaData: &silapb.SilaData{
				DepositRoot: make([]byte, 32),
				BlockHash:   make([]byte, 32),
			},
			Graffiti: make([]byte, 32),
			SyncAggregate: &silapb.SyncAggregate{
				SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
				SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
			},
			SilaPayload: &silaenginev1.SilaPayloadDeneb{
				ParentHash:    make([]byte, 32),
				FeeRecipient:  make([]byte, 20),
				StateRoot:     make([]byte, 32),
				ReceiptsRoot:  make([]byte, 32),
				LogsBloom:     make([]byte, 256),
				PrevRandao:    make([]byte, 32),
				ExtraData:     make([]byte, 0),
				BaseFeePerGas: make([]byte, 32),
				BlockHash:     make([]byte, 32),
				Transactions:  make([][]byte, 0),
				Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
			},
			BlsToSilaChanges: make([]*silapb.SignedBLSToSilaChange, 0),
			BlobKzgCommitments:    make([][]byte, 0),
			SilaRequests: &silaenginev1.SilaRequests{
				Deposits:       make([]*silaenginev1.DepositRequest, 0),
				Withdrawals:    make([]*silaenginev1.WithdrawalRequest, 0),
				Consolidations: make([]*silaenginev1.ConsolidationRequest, 0),
			},
		}
	case version.Fulu:
		body = &silapb.BeaconBlockBodyElectra{
			RandaoReveal: make([]byte, 96),
			SilaData: &silapb.SilaData{
				DepositRoot: make([]byte, 32),
				BlockHash:   make([]byte, 32),
			},
			Graffiti: make([]byte, 32),
			SyncAggregate: &silapb.SyncAggregate{
				SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
				SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
			},
			SilaPayload: &silaenginev1.SilaPayloadDeneb{
				ParentHash:    make([]byte, 32),
				FeeRecipient:  make([]byte, 20),
				StateRoot:     make([]byte, 32),
				ReceiptsRoot:  make([]byte, 32),
				LogsBloom:     make([]byte, 256),
				PrevRandao:    make([]byte, 32),
				ExtraData:     make([]byte, 0),
				BaseFeePerGas: make([]byte, 32),
				BlockHash:     make([]byte, 32),
				Transactions:  make([][]byte, 0),
				Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
			},
			BlsToSilaChanges: make([]*silapb.SignedBLSToSilaChange, 0),
			BlobKzgCommitments:    make([][]byte, 0),
			SilaRequests: &silaenginev1.SilaRequests{
				Deposits:       make([]*silaenginev1.DepositRequest, 0),
				Withdrawals:    make([]*silaenginev1.WithdrawalRequest, 0),
				Consolidations: make([]*silaenginev1.ConsolidationRequest, 0),
			},
		}
	default:
		return errUnsupportedVersion
	}

	root, err := body.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "could not hash tree root empty block body")
	}
	lbh := &silapb.BeaconBlockHeader{
		ParentRoot: params.BeaconConfig().ZeroHash[:],
		StateRoot:  params.BeaconConfig().ZeroHash[:],
		BodyRoot:   root[:],
	}
	return g.SetLatestBlockHeader(lbh)
}

func (s *PremineGenesisConfig) setSilaPayload(g state.BeaconState) error {
	gb := s.GB
	extraData := gb.Extra()
	if len(extraData) > 32 {
		extraData = extraData[:32]
	}

	if s.Version >= version.Deneb {
		payload := &silaenginev1.SilaPayloadDeneb{
			ParentHash:    gb.ParentHash().Bytes(),
			FeeRecipient:  gb.Coinbase().Bytes(),
			StateRoot:     gb.Root().Bytes(),
			ReceiptsRoot:  gb.ReceiptHash().Bytes(),
			LogsBloom:     gb.Bloom().Bytes(),
			PrevRandao:    params.BeaconConfig().ZeroHash[:],
			BlockNumber:   gb.NumberU64(),
			GasLimit:      gb.GasLimit(),
			GasUsed:       gb.GasUsed(),
			Timestamp:     gb.Time(),
			ExtraData:     extraData,
			BaseFeePerGas: bytesutil.PadTo(bytesutil.ReverseByteOrder(gb.BaseFee().Bytes()), fieldparams.RootLength),
			BlockHash:     gb.Hash().Bytes(),
			Transactions:  make([][]byte, 0),
			Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
			ExcessBlobGas: unwrapUint64Ptr(gb.ExcessBlobGas()),
			BlobGasUsed:   unwrapUint64Ptr(gb.BlobGasUsed()),
		}

		wep, err := blocks.WrappedSilaPayloadDeneb(payload)
		if err != nil {
			return err
		}

		eph, err := blocks.PayloadToHeaderDeneb(wep)
		if err != nil {
			return err
		}

		ed, err := blocks.WrappedSilaPayloadHeaderDeneb(eph)
		if err != nil {
			return err
		}

		return g.SetLatestSilaPayloadHeader(ed)
	}

	if s.Version >= version.Capella {
		payload := &silaenginev1.SilaPayloadCapella{
			ParentHash:    gb.ParentHash().Bytes(),
			FeeRecipient:  gb.Coinbase().Bytes(),
			StateRoot:     gb.Root().Bytes(),
			ReceiptsRoot:  gb.ReceiptHash().Bytes(),
			LogsBloom:     gb.Bloom().Bytes(),
			PrevRandao:    params.BeaconConfig().ZeroHash[:],
			BlockNumber:   gb.NumberU64(),
			GasLimit:      gb.GasLimit(),
			GasUsed:       gb.GasUsed(),
			Timestamp:     gb.Time(),
			ExtraData:     extraData,
			BaseFeePerGas: bytesutil.PadTo(bytesutil.ReverseByteOrder(gb.BaseFee().Bytes()), fieldparams.RootLength),
			BlockHash:     gb.Hash().Bytes(),
			Transactions:  make([][]byte, 0),
			Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
		}

		wep, err := blocks.WrappedSilaPayloadCapella(payload)
		if err != nil {
			return err
		}

		eph, err := blocks.PayloadToHeaderCapella(wep)
		if err != nil {
			return err
		}
		ed, err := blocks.WrappedSilaPayloadHeaderCapella(eph)
		if err != nil {
			return err
		}

		return g.SetLatestSilaPayloadHeader(ed)
	}

	if s.Version >= version.Bellatrix {
		payload := &silaenginev1.SilaPayload{
			ParentHash:    gb.ParentHash().Bytes(),
			FeeRecipient:  gb.Coinbase().Bytes(),
			StateRoot:     gb.Root().Bytes(),
			ReceiptsRoot:  gb.ReceiptHash().Bytes(),
			LogsBloom:     gb.Bloom().Bytes(),
			PrevRandao:    params.BeaconConfig().ZeroHash[:],
			BlockNumber:   gb.NumberU64(),
			GasLimit:      gb.GasLimit(),
			GasUsed:       gb.GasUsed(),
			Timestamp:     gb.Time(),
			ExtraData:     extraData,
			BaseFeePerGas: bytesutil.PadTo(bytesutil.ReverseByteOrder(gb.BaseFee().Bytes()), fieldparams.RootLength),
			BlockHash:     gb.Hash().Bytes(),
			Transactions:  make([][]byte, 0),
		}

		wep, err := blocks.WrappedSilaPayload(payload)
		if err != nil {
			return err
		}

		eph, err := blocks.PayloadToHeader(wep)
		if err != nil {
			return err
		}

		ed, err := blocks.WrappedSilaPayloadHeader(eph)
		if err != nil {
			return err
		}

		return g.SetLatestSilaPayloadHeader(ed)
	}

	if s.Version >= version.Phase0 {
		return nil
	}

	return errUnsupportedVersion
}

func unwrapUint64Ptr(u *uint64) uint64 {
	if u == nil {
		return 0
	}
	return *u
}

func nZeroRoots(n uint64) [][]byte {
	roots := make([][]byte, n)
	zh := params.BeaconConfig().ZeroHash[:]
	for i := range n {
		roots[i] = zh
	}
	return roots
}

func nSetRoots(n uint64, r []byte) [][]byte {
	roots := make([][]byte, n)
	for i := range n {
		h := make([]byte, 32)
		copy(h, r)
		roots[i] = h
	}
	return roots
}
