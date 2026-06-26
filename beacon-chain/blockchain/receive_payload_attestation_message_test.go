package blockchain

import (
	"testing"

	mockSila "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func TestReceivePayloadAttestationMessage_NilMessage(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	err := s.ReceivePayloadAttestationMessage(t.Context(), nil)
	require.ErrorContains(t, "nil payload attestation message", err)
}

func TestReceivePayloadAttestationMessage_NilData(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	msg := &silapb.PayloadAttestationMessage{}
	err := s.ReceivePayloadAttestationMessage(t.Context(), msg)
	require.ErrorContains(t, "nil payload attestation message", err)
}

func TestReceivePayloadAttestationMessage_ValidatorNotInPTC(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig()
	cfg.GloasForkEpoch = 0
	params.OverrideBeaconConfig(cfg)

	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	ctx := t.Context()

	blockRoot := bytesutil.ToBytes32([]byte("root1"))
	parentRoot := params.BeaconConfig().ZeroHash
	blockHash := bytesutil.ToBytes32([]byte("hash1"))

	numVals := 2048
	headState := gloasStateWithValidators(t, 1, numVals)

	base, blk := testGloasState(t, 1, parentRoot, blockHash)
	insertGloasBlock(t, s, base, blk, blockRoot)

	wsb, err := blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	s.head = &head{root: blockRoot, block: wsb, state: headState, slot: 1}

	ptc, err := headState.PayloadCommitteeReadOnly(1)
	require.NoError(t, err)

	// Pick a validator index not in the PTC.
	inPTC := make(map[primitives.ValidatorIndex]bool)
	for _, idx := range ptc {
		inPTC[idx] = true
	}
	var notInPTC primitives.ValidatorIndex
	for i := primitives.ValidatorIndex(0); int(i) < numVals; i++ {
		if !inPTC[i] {
			notInPTC = i
			break
		}
	}

	msg := &silapb.PayloadAttestationMessage{
		ValidatorIndex: notInPTC,
		Data: &silapb.PayloadAttestationData{
			BeaconBlockRoot: blockRoot[:],
			Slot:            1,
		},
	}
	err = s.ReceivePayloadAttestationMessage(ctx, msg)
	require.ErrorContains(t, "validator not in PTC", err)
}

func TestReceivePayloadAttestationMessage_OK(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig()
	cfg.GloasForkEpoch = 0
	params.OverrideBeaconConfig(cfg)

	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	ctx := t.Context()

	blockRoot := bytesutil.ToBytes32([]byte("root1"))
	parentRoot := params.BeaconConfig().ZeroHash
	blockHash := bytesutil.ToBytes32([]byte("hash1"))

	headState := gloasStateWithValidators(t, 1, 2048)

	base, blk := testGloasState(t, 1, parentRoot, blockHash)
	insertGloasBlock(t, s, base, blk, blockRoot)

	wsb, err := blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	s.head = &head{root: blockRoot, block: wsb, state: headState, slot: 1}

	ptc, err := headState.PayloadCommitteeReadOnly(1)
	require.NoError(t, err)
	require.NotEqual(t, 0, len(ptc))

	msg := &silapb.PayloadAttestationMessage{
		ValidatorIndex: ptc[0],
		Data: &silapb.PayloadAttestationData{
			BeaconBlockRoot:   blockRoot[:],
			Slot:              1,
			PayloadPresent:    true,
			BlobDataAvailable: true,
		},
	}
	require.NoError(t, s.ReceivePayloadAttestationMessage(ctx, msg))
}

// gloasStateWithValidators returns a Gloas beacon state with active validators
// for PTC committee computation.
func gloasStateWithValidators(t *testing.T, slot primitives.Slot, numVals int) state.BeaconState {
	t.Helper()
	validators := make([]*silapb.Validator, numVals)
	balances := make([]uint64, numVals)
	for i := range validators {
		validators[i] = &silapb.Validator{
			PublicKey:             make([]byte, 48),
			WithdrawalCredentials: make([]byte, 32),
			EffectiveBalance:      params.BeaconConfig().MaxEffectiveBalanceElectra,
			ActivationEpoch:       0,
			ExitEpoch:             params.BeaconConfig().FarFutureEpoch,
		}
		balances[i] = params.BeaconConfig().MaxEffectiveBalanceElectra
	}
	st, err := util.NewBeaconStateGloas(func(s *silapb.BeaconStateGloas) error {
		s.Slot = slot
		s.Validators = validators
		s.Balances = balances
		return nil
	})
	require.NoError(t, err)
	return st
}
