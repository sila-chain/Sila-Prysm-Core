package blocks

import (
	"testing"

	bitfield "github.com/sila-chain/go-bitfield"
	consensus_types "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	eth "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func TestSignedBeaconBlock_SetPayloadAttestations(t *testing.T) {
	t.Run("rejects pre-Gloas versions", func(t *testing.T) {
		sb := newTestSignedBeaconBlock(version.Fulu)
		payload := []*eth.PayloadAttestation{{}}

		err := sb.SetPayloadAttestations(payload)

		require.ErrorIs(t, err, consensus_types.ErrUnsupportedField)
		require.IsNil(t, sb.block.body.payloadAttestations)
	})

	t.Run("sets payload attestations for Gloas", func(t *testing.T) {
		sb := newTestSignedBeaconBlock(version.Gloas)
		payload := []*eth.PayloadAttestation{
			{
				AggregationBits: bitfield.NewBitvector512(),
				Data: &eth.PayloadAttestationData{
					BeaconBlockRoot:   []byte{0x01, 0x02},
					PayloadPresent:    true,
					BlobDataAvailable: true,
				},
				Signature: []byte{0x03},
			},
		}

		err := sb.SetPayloadAttestations(payload)

		require.NoError(t, err)
		require.DeepEqual(t, payload, sb.block.body.payloadAttestations)
	})
}

func TestSignedBeaconBlock_SetSignedSilaPayloadBid(t *testing.T) {
	t.Run("rejects pre-Gloas versions", func(t *testing.T) {
		sb := newTestSignedBeaconBlock(version.Fulu)
		payloadBid := &eth.SignedSilaPayloadBid{}

		err := sb.SetSignedSilaPayloadBid(payloadBid)

		require.ErrorIs(t, err, consensus_types.ErrUnsupportedField)
		require.IsNil(t, sb.block.body.signedSilaPayloadBid)
	})

	t.Run("sets signed sila payload bid for Gloas", func(t *testing.T) {
		sb := newTestSignedBeaconBlock(version.Gloas)
		payloadBid := &eth.SignedSilaPayloadBid{
			Message: &eth.SilaPayloadBid{
				ParentBlockHash: []byte{0xaa},
				BlockHash:       []byte{0xbb},
				FeeRecipient:    []byte{0xcc},
			},
			Signature: []byte{0xdd},
		}

		err := sb.SetSignedSilaPayloadBid(payloadBid)

		require.NoError(t, err)
		require.Equal(t, payloadBid, sb.block.body.signedSilaPayloadBid)
	})
}

func TestSignedBeaconBlock_SetExecution(t *testing.T) {
	t.Run("rejects Gloas version", func(t *testing.T) {
		sb := newTestSignedBeaconBlock(version.Gloas)
		payload := &silaenginev1.SilaPayload{}
		wrapped, err := WrappedSilaPayload(payload)
		require.NoError(t, err)

		err = sb.SetExecution(wrapped)
		require.ErrorIs(t, err, consensus_types.ErrUnsupportedField)
	})
}

func TestSignedBeaconBlock_SetExecutionRequests(t *testing.T) {
	t.Run("rejects Gloas version", func(t *testing.T) {
		sb := newTestSignedBeaconBlock(version.Gloas)
		requests := &silaenginev1.ExecutionRequests{}

		err := sb.SetExecutionRequests(requests)
		require.ErrorIs(t, err, consensus_types.ErrUnsupportedField)
	})
}

func newTestSignedBeaconBlock(ver int) *SignedBeaconBlock {
	return &SignedBeaconBlock{
		version: ver,
		block: &BeaconBlock{
			version: ver,
			body: &BeaconBlockBody{
				version: ver,
			},
		},
	}
}
