package gloas

import (
	"bytes"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"google.golang.org/protobuf/proto"
)

type payloadFixture struct {
	state       state.BeaconState
	signed      interfaces.ROSignedSilaPayloadEnvelope
	signedProto *silapb.SignedSilaPayloadEnvelope
	envelope    *silapb.SilaPayloadEnvelope
	payload     *silaenginev1.SilaPayloadGloas
	slot        primitives.Slot
}

func buildPayloadFixture(t *testing.T, mutate func(payload *silaenginev1.SilaPayloadGloas, bid *silapb.SilaPayloadBid, envelope *silapb.SilaPayloadEnvelope)) payloadFixture {
	t.Helper()

	cfg := params.BeaconConfig()
	slot := primitives.Slot(5)
	builderIdx := primitives.BuilderIndex(0)

	sk, err := bls.RandKey()
	require.NoError(t, err)
	pk := sk.PublicKey().Marshal()

	randao := bytes.Repeat([]byte{0xAA}, 32)
	parentHash := bytes.Repeat([]byte{0xBB}, 32)
	blockHash := bytes.Repeat([]byte{0xCC}, 32)

	withdrawals := []*silaenginev1.Withdrawal{
		{Index: 0, ValidatorIndex: 1, Address: bytes.Repeat([]byte{0x01}, 20), Amount: 0},
	}

	payload := &silaenginev1.SilaPayloadGloas{
		ParentHash:    parentHash,
		FeeRecipient:  bytes.Repeat([]byte{0x01}, 20),
		StateRoot:     bytes.Repeat([]byte{0x02}, 32),
		ReceiptsRoot:  bytes.Repeat([]byte{0x03}, 32),
		LogsBloom:     bytes.Repeat([]byte{0x04}, 256),
		PrevRandao:    randao,
		BlockNumber:   1,
		GasLimit:      1,
		GasUsed:       0,
		Timestamp:     100,
		ExtraData:     []byte{},
		BaseFeePerGas: bytes.Repeat([]byte{0x05}, 32),
		BlockHash:     blockHash,
		Transactions:  [][]byte{},
		Withdrawals:   withdrawals,
		BlobGasUsed:   0,
		ExcessBlobGas: 0,
		SlotNumber:    slot,
	}

	emptyRequestsRoot, err := silaenginev1.EmptyExecutionRequestsHashTreeRoot()
	require.NoError(t, err)
	bid := &silapb.SilaPayloadBid{
		ParentBlockHash:       parentHash,
		ParentBlockRoot:       bytes.Repeat([]byte{0xDD}, 32),
		BlockHash:             blockHash,
		PrevRandao:            randao,
		GasLimit:              1,
		BuilderIndex:          builderIdx,
		Slot:                  slot,
		Value:                 0,
		ExecutionPayment:      0,
		FeeRecipient:          bytes.Repeat([]byte{0xEE}, 20),
		ExecutionRequestsRoot: emptyRequestsRoot[:],
	}

	header := &silapb.BeaconBlockHeader{
		Slot:       slot,
		ParentRoot: bytes.Repeat([]byte{0x11}, 32),
		StateRoot:  bytes.Repeat([]byte{0x22}, 32),
		BodyRoot:   bytes.Repeat([]byte{0x33}, 32),
	}
	headerRoot, err := header.HashTreeRoot()
	require.NoError(t, err)

	envelope := &silapb.SilaPayloadEnvelope{
		BuilderIndex:          builderIdx,
		BeaconBlockRoot:       headerRoot[:],
		ParentBeaconBlockRoot: header.ParentRoot,
		Payload:               payload,
		ExecutionRequests:     &silaenginev1.ExecutionRequests{},
	}

	if mutate != nil {
		mutate(payload, bid, envelope)
	}

	genesisRoot := bytes.Repeat([]byte{0xAB}, 32)
	blockRoots := make([][]byte, cfg.SlotsPerHistoricalRoot)
	stateRoots := make([][]byte, cfg.SlotsPerHistoricalRoot)
	for i := range blockRoots {
		blockRoots[i] = bytes.Repeat([]byte{0x44}, 32)
		stateRoots[i] = bytes.Repeat([]byte{0x55}, 32)
	}
	randaoMixes := make([][]byte, cfg.EpochsPerHistoricalVector)
	for i := range randaoMixes {
		randaoMixes[i] = randao
	}

	withdrawalCreds := make([]byte, 32)
	withdrawalCreds[0] = cfg.SilaExecutionAddressWithdrawalPrefixByte

	silaexecData := &silapb.SilaExecutionData{
		DepositRoot:  bytes.Repeat([]byte{0x66}, 32),
		DepositCount: 0,
		BlockHash:    bytes.Repeat([]byte{0x77}, 32),
	}

	vals := []*silapb.Validator{
		{
			PublicKey:             pk,
			WithdrawalCredentials: withdrawalCreds,
			EffectiveBalance:      cfg.MinActivationBalance + 1_000,
		},
	}
	balances := []uint64{cfg.MinActivationBalance + 1_000}

	payments := make([]*silapb.BuilderPendingPayment, cfg.SlotsPerEpoch*2)
	for i := range payments {
		payments[i] = &silapb.BuilderPendingPayment{
			Withdrawal: &silapb.BuilderPendingWithdrawal{
				FeeRecipient: make([]byte, 20),
			},
		}
	}

	silaPayloadAvailability := make([]byte, cfg.SlotsPerHistoricalRoot/8)

	builders := make([]*silapb.Builder, builderIdx+1)
	builders[builderIdx] = &silapb.Builder{
		Pubkey:            pk,
		Version:           []byte{0},
		ExecutionAddress:  bytes.Repeat([]byte{0x09}, 20),
		Balance:           0,
		DepositEpoch:      0,
		WithdrawableEpoch: 0,
	}

	genesisTime := uint64(0)
	slotSeconds := cfg.SecondsPerSlot * uint64(slot)
	if payload.Timestamp > slotSeconds {
		genesisTime = payload.Timestamp - slotSeconds
	}

	stProto := &silapb.BeaconStateGloas{
		Slot:                  slot,
		GenesisTime:           genesisTime,
		GenesisValidatorsRoot: genesisRoot,
		Fork: &silapb.Fork{
			CurrentVersion:  bytes.Repeat([]byte{0x01}, 4),
			PreviousVersion: bytes.Repeat([]byte{0x01}, 4),
			Epoch:           0,
		},
		LatestBlockHeader:            header,
		BlockRoots:                   blockRoots,
		StateRoots:                   stateRoots,
		RandaoMixes:                  randaoMixes,
		SilaExecutionData:                     silaexecData,
		Validators:                   vals,
		Balances:                     balances,
		LatestBlockHash:              payload.ParentHash,
		LatestSilaPayloadBid:    bid,
		BuilderPendingPayments:       payments,
		SilaPayloadAvailability: silaPayloadAvailability,
		BuilderPendingWithdrawals:    []*silapb.BuilderPendingWithdrawal{},
		PayloadExpectedWithdrawals:   payload.Withdrawals,
		Builders:                     builders,
	}

	st, err := state_native.InitializeFromProtoGloas(stProto)
	require.NoError(t, err)

	epoch := slots.ToEpoch(slot)
	domain, err := signing.Domain(st.Fork(), epoch, cfg.DomainBeaconBuilder, st.GenesisValidatorsRoot())
	require.NoError(t, err)
	signingRoot, err := signing.ComputeSigningRoot(envelope, domain)
	require.NoError(t, err)
	signature := sk.Sign(signingRoot[:]).Marshal()

	signedProto := &silapb.SignedSilaPayloadEnvelope{
		Message:   envelope,
		Signature: signature,
	}
	signed, err := blocks.WrappedROSignedSilaPayloadEnvelope(signedProto)
	require.NoError(t, err)

	return payloadFixture{
		state:       st,
		signed:      signed,
		signedProto: signedProto,
		envelope:    envelope,
		payload:     payload,
		slot:        slot,
	}
}

func TestVerifySilaPayloadEnvelope_Success(t *testing.T) {
	fixture := buildPayloadFixture(t, nil)
	require.NoError(t, VerifySilaPayloadEnvelope(t.Context(), fixture.state, fixture.signed))
}

func TestVerifySilaPayloadEnvelopeWithDeferredSig_Success(t *testing.T) {
	fixture := buildPayloadFixture(t, nil)

	sigBatch, err := VerifySilaPayloadEnvelopeWithDeferredSig(t.Context(), fixture.state, fixture.signed)
	require.NoError(t, err)
	require.NotNil(t, sigBatch)
	require.Equal(t, 1, len(sigBatch.Signatures))
	require.Equal(t, 1, len(sigBatch.PublicKeys))
	require.Equal(t, 1, len(sigBatch.Messages))
	require.Equal(t, 1, len(sigBatch.Descriptions))
	require.Equal(t, "sila payload envelope signature", sigBatch.Descriptions[0])

	valid, err := sigBatch.Verify()
	require.NoError(t, err)
	require.Equal(t, true, valid)
}

func TestVerifySilaPayloadEnvelopeSignature(t *testing.T) {
	fixture := buildPayloadFixture(t, nil)

	t.Run("self build", func(t *testing.T) {
		proposerSk, err := bls.RandKey()
		require.NoError(t, err)
		proposerPk := proposerSk.PublicKey().Marshal()

		stPb, ok := fixture.state.ToProtoUnsafe().(*silapb.BeaconStateGloas)
		require.Equal(t, true, ok)
		stPb = proto.Clone(stPb).(*silapb.BeaconStateGloas)
		stPb.Validators[0].PublicKey = proposerPk
		st, err := state_native.InitializeFromProtoUnsafeGloas(stPb)
		require.NoError(t, err)

		msg := proto.Clone(fixture.signedProto.Message).(*silapb.SilaPayloadEnvelope)
		msg.BuilderIndex = params.BeaconConfig().BuilderIndexSelfBuild

		epoch := slots.ToEpoch(msg.Payload.SlotNumber)
		domain, err := signing.Domain(st.Fork(), epoch, params.BeaconConfig().DomainBeaconBuilder, st.GenesisValidatorsRoot())
		require.NoError(t, err)
		signingRoot, err := signing.ComputeSigningRoot(msg, domain)
		require.NoError(t, err)
		signature := proposerSk.Sign(signingRoot[:]).Marshal()

		signedProto := &silapb.SignedSilaPayloadEnvelope{
			Message:   msg,
			Signature: signature,
		}
		signed, err := blocks.WrappedROSignedSilaPayloadEnvelope(signedProto)
		require.NoError(t, err)

		require.NoError(t, verifySilaPayloadEnvelopeSignature(st, signed))
	})

	t.Run("builder", func(t *testing.T) {
		signed, err := blocks.WrappedROSignedSilaPayloadEnvelope(fixture.signedProto)
		require.NoError(t, err)

		require.NoError(t, verifySilaPayloadEnvelopeSignature(fixture.state, signed))
	})

	t.Run("invalid signature", func(t *testing.T) {
		t.Run("self build", func(t *testing.T) {
			proposerSk, err := bls.RandKey()
			require.NoError(t, err)
			proposerPk := proposerSk.PublicKey().Marshal()

			stPb, ok := fixture.state.ToProtoUnsafe().(*silapb.BeaconStateGloas)
			require.Equal(t, true, ok)
			stPb = proto.Clone(stPb).(*silapb.BeaconStateGloas)
			stPb.Validators[0].PublicKey = proposerPk
			st, err := state_native.InitializeFromProtoUnsafeGloas(stPb)
			require.NoError(t, err)

			msg := proto.Clone(fixture.signedProto.Message).(*silapb.SilaPayloadEnvelope)
			msg.BuilderIndex = params.BeaconConfig().BuilderIndexSelfBuild

			signedProto := &silapb.SignedSilaPayloadEnvelope{
				Message:   msg,
				Signature: bytes.Repeat([]byte{0xFF}, 96),
			}
			badSigned, err := blocks.WrappedROSignedSilaPayloadEnvelope(signedProto)
			require.NoError(t, err)

			err = verifySilaPayloadEnvelopeSignature(st, badSigned)
			require.ErrorContains(t, "invalid signature format", err)
		})

		t.Run("builder", func(t *testing.T) {
			signedProto := &silapb.SignedSilaPayloadEnvelope{
				Message:   fixture.signedProto.Message,
				Signature: bytes.Repeat([]byte{0xFF}, 96),
			}
			badSigned, err := blocks.WrappedROSignedSilaPayloadEnvelope(signedProto)
			require.NoError(t, err)

			err = verifySilaPayloadEnvelopeSignature(fixture.state, badSigned)
			require.ErrorContains(t, "invalid signature format", err)
		})
	})
}
