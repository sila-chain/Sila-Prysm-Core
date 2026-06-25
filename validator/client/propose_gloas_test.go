package client

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	consensusblocks "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"github.com/pkg/errors"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
)

func signedGloasBlock(t *testing.T, slot primitives.Slot, builderIndex primitives.BuilderIndex) interfaces.SignedBeaconBlock {
	t.Helper()

	blk := util.NewBeaconBlockGloas()
	blk.Block.Slot = slot
	if blk.Block.Body == nil {
		blk.Block.Body = &silapb.BeaconBlockBodyGloas{}
	}
	blk.Block.Body.SignedSilaPayloadBid = util.HydrateSignedSilaPayloadBid(&silapb.SignedSilaPayloadBid{
		Message: &silapb.SilaPayloadBid{
			BuilderIndex: builderIndex,
		},
		Signature: make([]byte, 96),
	})

	signed, err := consensusblocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	return signed
}

func testSilaPayloadEnvelope(slot primitives.Slot, builderIndex primitives.BuilderIndex) *silapb.SilaPayloadEnvelope {
	return &silapb.SilaPayloadEnvelope{
		Payload: &silaenginev1.SilaPayloadGloas{
			ParentHash:    make([]byte, 32),
			FeeRecipient:  make([]byte, 20),
			StateRoot:     make([]byte, 32),
			ReceiptsRoot:  make([]byte, 32),
			LogsBloom:     make([]byte, 256),
			PrevRandao:    make([]byte, 32),
			BaseFeePerGas: make([]byte, 32),
			BlockHash:     make([]byte, 32),
			ExtraData:     make([]byte, 0),
			SlotNumber:    slot,
		},
		ExecutionRequests:     &silaenginev1.ExecutionRequests{},
		BuilderIndex:          builderIndex,
		BeaconBlockRoot:       make([]byte, 32),
		ParentBeaconBlockRoot: make([]byte, 32),
	}
}

func TestProposeSelfBuildEnvelope(t *testing.T) {
	validator, m, validatorKey, finish := setup(t, false)
	defer finish()

	slot := primitives.Slot(100)
	builderIndex := params.BeaconConfig().BuilderIndexSelfBuild

	expectedEnvelope := testSilaPayloadEnvelope(slot, builderIndex)

	m.validatorClient.EXPECT().
		GetSilaPayloadEnvelope(gomock.Any(), slot, gomock.Any()).
		Return(expectedEnvelope, nil, nil)

	builderDomain := make([]byte, 32)
	copy(builderDomain, params.BeaconConfig().DomainBeaconBuilder[:])
	m.validatorClient.EXPECT().
		DomainData(gomock.Any(), gomock.Any()).
		Return(&silapb.DomainResponse{SignatureDomain: builderDomain}, nil)

	m.validatorClient.EXPECT().
		PublishSilaPayloadEnvelope(gomock.Any(), gomock.AssignableToTypeOf(&silapb.SignedSilaPayloadEnvelope{})).
		Return(&emptypb.Empty{}, nil)

	signedBlock := signedGloasBlock(t, slot, builderIndex)

	var pubKey [fieldparams.BLSPubkeyLength]byte
	copy(pubKey[:], validatorKey.PublicKey().Marshal())

	err := validator.proposeSelfBuildEnvelope(t.Context(), slot, pubKey, signedBlock)
	require.NoError(t, err)
}

func TestProposeSelfBuildEnvelope_MissingBid(t *testing.T) {
	validator, _, _, finish := setup(t, false)
	defer finish()

	blk := util.NewBeaconBlockGloas()
	blk.Block.Slot = 1
	if blk.Block.Body == nil {
		blk.Block.Body = &silapb.BeaconBlockBodyGloas{}
	}
	blk.Block.Body.SignedSilaPayloadBid = nil

	signedBlock, err := consensusblocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)

	var pubKey [fieldparams.BLSPubkeyLength]byte
	err = validator.proposeSelfBuildEnvelope(t.Context(), 1, pubKey, signedBlock)
	require.ErrorContains(t, "no sila payload bid found in block body", err)
}

func TestProposeSelfBuildEnvelope_ClientError(t *testing.T) {
	validator, m, validatorKey, finish := setup(t, false)
	defer finish()

	m.validatorClient.EXPECT().
		GetSilaPayloadEnvelope(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil, errors.New("connection refused"))

	signedBlock := signedGloasBlock(t, 1, params.BeaconConfig().BuilderIndexSelfBuild)

	var pubKey [fieldparams.BLSPubkeyLength]byte
	copy(pubKey[:], validatorKey.PublicKey().Marshal())

	err := validator.proposeSelfBuildEnvelope(t.Context(), 1, pubKey, signedBlock)
	require.ErrorContains(t, "failed to get sila payload envelope for self-build", err)
}

func TestSignSilaPayloadEnvelope(t *testing.T) {
	validator, m, _, finish := setup(t, false)
	defer finish()

	kp := testKeyFromBytes(t, []byte{1})
	validator.km = newMockKeymanager(t, kp)

	builderDomain := make([]byte, 32)
	copy(builderDomain, params.BeaconConfig().DomainBeaconBuilder[:])
	m.validatorClient.EXPECT().
		DomainData(gomock.Any(), gomock.Any()).
		Return(&silapb.DomainResponse{SignatureDomain: builderDomain}, nil)

	envelope := testSilaPayloadEnvelope(100, 42)

	signed, err := validator.signSilaPayloadEnvelope(t.Context(), kp.pub, 100, envelope)
	require.NoError(t, err)
	require.NotNil(t, signed)
	require.DeepEqual(t, envelope, signed.Message)
	require.NotNil(t, signed.Signature)

	// Verify the signature was computed with the builder domain.
	expectedRoot, err := signing.ComputeSigningRoot(envelope, builderDomain)
	require.NoError(t, err)
	require.NotEqual(t, [32]byte{}, expectedRoot)
}

func TestSignSilaPayloadEnvelope_VerifySignature(t *testing.T) {
	validator, m, _, finish := setup(t, false)
	defer finish()

	kp := testKeyFromBytes(t, []byte{1})
	validator.km = newMockKeymanager(t, kp)

	builderDomain := make([]byte, 32)
	copy(builderDomain, params.BeaconConfig().DomainBeaconBuilder[:])
	m.validatorClient.EXPECT().
		DomainData(gomock.Any(), gomock.Any()).
		Return(&silapb.DomainResponse{SignatureDomain: builderDomain}, nil)

	envelope := testSilaPayloadEnvelope(100, 42)

	signed, err := validator.signSilaPayloadEnvelope(t.Context(), kp.pub, 100, envelope)
	require.NoError(t, err)

	// Compute the expected signing root and verify the signature.
	signingRoot, err := signing.ComputeSigningRoot(envelope, builderDomain)
	require.NoError(t, err)

	sig, err := bls.SignatureFromBytes(signed.Signature)
	require.NoError(t, err)
	require.Equal(t, true, sig.Verify(kp.pri.PublicKey(), signingRoot[:]))
}

func TestSignSilaPayloadEnvelope_DomainDataError(t *testing.T) {
	validator, m, _, finish := setup(t, false)
	defer finish()

	kp := testKeyFromBytes(t, []byte{1})
	validator.km = newMockKeymanager(t, kp)

	m.validatorClient.EXPECT().
		DomainData(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("domain data unavailable"))

	envelope := testSilaPayloadEnvelope(100, 0)

	_, err := validator.signSilaPayloadEnvelope(t.Context(), kp.pub, 100, envelope)
	require.ErrorContains(t, "could not get domain data", err)
}

func TestSignSilaPayloadEnvelope_NilDomain(t *testing.T) {
	validator, m, _, finish := setup(t, false)
	defer finish()

	kp := testKeyFromBytes(t, []byte{1})
	validator.km = newMockKeymanager(t, kp)

	m.validatorClient.EXPECT().
		DomainData(gomock.Any(), gomock.Any()).
		Return(nil, nil)

	envelope := testSilaPayloadEnvelope(100, 0)

	_, err := validator.signSilaPayloadEnvelope(t.Context(), kp.pub, 100, envelope)
	require.ErrorContains(t, "nil domain data", err)
}

func TestSignSilaPayloadEnvelope_UsesDomainBeaconBuilder(t *testing.T) {
	validator, m, _, finish := setup(t, false)
	defer finish()

	kp := testKeyFromBytes(t, []byte{1})
	validator.km = newMockKeymanager(t, kp)

	// Verify the correct domain type is requested.
	m.validatorClient.EXPECT().
		DomainData(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx any, req *silapb.DomainRequest) (*silapb.DomainResponse, error) {
			require.DeepEqual(t, params.BeaconConfig().DomainBeaconBuilder[:], req.Domain)
			return &silapb.DomainResponse{SignatureDomain: make([]byte, 32)}, nil
		})

	envelope := testSilaPayloadEnvelope(100, 0)

	_, err := validator.signSilaPayloadEnvelope(t.Context(), kp.pub, 100, envelope)
	require.NoError(t, err)
}

// TestProposeBlock_Gloas_EnvelopeAfterBlock verifies that the Gloas propose flow
// submits the block first, then retrieves, signs, and publishes the envelope.
// The envelope's state root is lazily computed by the beacon node from the
// post-block state, so this ordering is critical.
func TestProposeBlock_Gloas_EnvelopeAfterBlock(t *testing.T) {
	hook := logTest.NewGlobal()
	validator, m, validatorKey, finish := setup(t, false)
	defer finish()

	var pubKey [fieldparams.BLSPubkeyLength]byte
	copy(pubKey[:], validatorKey.PublicKey().Marshal())

	blk := util.NewBeaconBlockGloas()
	builderIndex := params.BeaconConfig().BuilderIndexSelfBuild
	if blk.Block.Body == nil {
		blk.Block.Body = &silapb.BeaconBlockBodyGloas{}
	}
	if blk.Block.Body.SignedSilaPayloadBid == nil {
		blk.Block.Body.SignedSilaPayloadBid = &silapb.SignedSilaPayloadBid{}
	}
	if blk.Block.Body.SignedSilaPayloadBid.Message == nil {
		blk.Block.Body.SignedSilaPayloadBid.Message = &silapb.SilaPayloadBid{}
	}
	blk.Block.Body.SignedSilaPayloadBid.Message.BuilderIndex = builderIndex

	gloasBlock := &silapb.GenericBeaconBlock{
		Block: &silapb.GenericBeaconBlock_Gloas{
			Gloas: blk.Block,
		},
	}

	envelope := testSilaPayloadEnvelope(1, builderIndex)

	// DomainData for randao signing.
	m.validatorClient.EXPECT().
		DomainData(gomock.Any(), gomock.Any()).
		Return(&silapb.DomainResponse{SignatureDomain: make([]byte, 32)}, nil)

	// BeaconBlock returns a Gloas block.
	m.validatorClient.EXPECT().
		BeaconBlock(gomock.Any(), gomock.AssignableToTypeOf(&silapb.BlockRequest{})).
		Return(gloasBlock, nil)

	// DomainData for block signing.
	m.validatorClient.EXPECT().
		DomainData(gomock.Any(), gomock.Any()).
		Return(&silapb.DomainResponse{SignatureDomain: make([]byte, 32)}, nil)

	// Critical ordering: ProposeBeaconBlock must be called BEFORE SilaPayloadEnvelope.
	proposeCall := m.validatorClient.EXPECT().
		ProposeBeaconBlock(gomock.Any(), gomock.AssignableToTypeOf(&silapb.GenericSignedBeaconBlock{})).
		Return(&silapb.ProposeResponse{BlockRoot: make([]byte, 32)}, nil)

	getEnvelopeCall := m.validatorClient.EXPECT().
		GetSilaPayloadEnvelope(gomock.Any(), primitives.Slot(1), gomock.Any()).
		Return(envelope, nil, nil).
		After(proposeCall)

	// DomainData for envelope signing.
	m.validatorClient.EXPECT().
		DomainData(gomock.Any(), gomock.Any()).
		Return(&silapb.DomainResponse{SignatureDomain: make([]byte, 32)}, nil).
		After(getEnvelopeCall)

	m.validatorClient.EXPECT().
		PublishSilaPayloadEnvelope(gomock.Any(), gomock.AssignableToTypeOf(&silapb.SignedSilaPayloadEnvelope{})).
		Return(&emptypb.Empty{}, nil)

	validator.ProposeBlock(t.Context(), 1, pubKey)
	require.LogsContain(t, hook, "Submitted new block")
}
