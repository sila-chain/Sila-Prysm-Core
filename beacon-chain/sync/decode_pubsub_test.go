package sync

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain"
	mock "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p"
	p2ptesting "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p/types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/startup"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/wrapper"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/metadata"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"github.com/d4l3k/messagediff"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
)

func TestService_decodePubsubMessage(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	params.BeaconConfig().InitializeForkSchedule()
	entry := params.GetNetworkScheduleEntry(params.BeaconConfig().GenesisEpoch)
	tests := []struct {
		name    string
		topic   string
		input   *pubsub.Message
		want    any
		wantErr error
	}{
		{
			name:    "Nil message",
			input:   nil,
			wantErr: errNilPubsubMessage,
		},
		{
			name: "nil topic",
			input: &pubsub.Message{
				Message: &pb.Message{
					Topic: nil,
				},
			},
			wantErr: errNilPubsubMessage,
		},
		{
			name:    "invalid topic format",
			topic:   "foo",
			wantErr: p2p.ErrInvalidTopic,
		},
		{
			name:    "topic not mapped to any message type",
			topic:   "/sila/abababab/foo/ssz_snappy",
			wantErr: p2p.ErrMessageNotMapped,
		},
		{
			name:  "valid message -- beacon block",
			topic: fmt.Sprintf(p2p.GossipTypeMapping[reflect.TypeFor[*silapb.SignedBeaconBlock]()], entry.ForkDigest),
			input: &pubsub.Message{
				Message: &pb.Message{
					Data: func() []byte {
						buf := new(bytes.Buffer)
						if _, err := p2ptesting.NewTestP2P(t).Encoding().EncodeGossip(buf, util.NewBeaconBlock()); err != nil {
							t.Fatal(err)
						}
						return buf.Bytes()
					}(),
				},
			},
			wantErr: nil,
			want: func() interfaces.ReadOnlySignedBeaconBlock {
				wsb, err := blocks.NewSignedBeaconBlock(util.NewBeaconBlock())
				require.NoError(t, err)
				return wsb
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := &mock.ChainService{ValidatorsRoot: [32]byte{}, Genesis: time.Now()}
			s := &Service{
				cfg: &config{p2p: p2ptesting.NewTestP2P(t), chain: chain, clock: startup.NewClock(chain.Genesis, chain.ValidatorsRoot)},
			}
			if tt.topic != "" {
				if tt.input == nil {
					tt.input = &pubsub.Message{Message: &pb.Message{}}
				} else if tt.input.Message == nil {
					tt.input.Message = &pb.Message{}
				}
				// reassign because tt is a loop variable
				topic := tt.topic
				tt.input.Message.Topic = &topic
			}
			got, err := s.decodePubsubMessage(tt.input)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr, "decodePubsubMessage() error mismatch")
				return
			}
			require.NoError(t, err, "decodePubsubMessage() unexpected error")
			if !reflect.DeepEqual(got, tt.want) {
				diff, _ := messagediff.PrettyDiff(got, tt.want)
				t.Log(diff)
				t.Errorf("decodePubsubMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractDataType(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	params.BeaconConfig().FuluForkEpoch = params.BeaconConfig().ElectraForkEpoch + 4096*2
	params.BeaconConfig().InitializeForkSchedule()

	type args struct {
		digest [4]byte
		chain  blockchain.ChainInfoFetcher
	}
	tests := []struct {
		name            string
		args            args
		wantBlock       interfaces.ReadOnlySignedBeaconBlock
		wantMd          metadata.Metadata
		wantAtt         silapb.Att
		wantAggregate   silapb.SignedAggregateAttAndProof
		wantAttSlashing silapb.AttSlashing
		wantErr         bool
	}{
		{
			name: "non existent digest",
			args: args{
				digest: [4]byte{},
				chain:  &mock.ChainService{ValidatorsRoot: [32]byte{}},
			},
			wantBlock:       nil,
			wantMd:          nil,
			wantAtt:         nil,
			wantAggregate:   nil,
			wantAttSlashing: nil,
			wantErr:         true,
		},
		{
			name: "genesis fork version",
			args: args{
				digest: params.ForkDigest(params.BeaconConfig().GenesisEpoch),
				chain:  &mock.ChainService{ValidatorsRoot: [32]byte{}},
			},
			wantBlock: func() interfaces.ReadOnlySignedBeaconBlock {
				wsb, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlock{Block: &silapb.BeaconBlock{Body: &silapb.BeaconBlockBody{}}})
				require.NoError(t, err)
				return wsb
			}(),
			wantAtt:         &silapb.Attestation{},
			wantAggregate:   &silapb.SignedAggregateAttestationAndProof{},
			wantAttSlashing: &silapb.AttesterSlashing{},
			wantErr:         false,
		},
		{
			name: "altair fork version",
			args: args{
				digest: params.ForkDigest(params.BeaconConfig().AltairForkEpoch),
				chain:  &mock.ChainService{ValidatorsRoot: [32]byte{}},
			},
			wantBlock: func() interfaces.ReadOnlySignedBeaconBlock {
				wsb, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockAltair{Block: &silapb.BeaconBlockAltair{Body: &silapb.BeaconBlockBodyAltair{}}})
				require.NoError(t, err)
				return wsb
			}(),
			wantMd:          wrapper.WrappedMetadataV1(&silapb.MetaDataV1{}),
			wantAtt:         &silapb.Attestation{},
			wantAggregate:   &silapb.SignedAggregateAttestationAndProof{},
			wantAttSlashing: &silapb.AttesterSlashing{},
			wantErr:         false,
		},
		{
			name: "bellatrix fork version",
			args: args{
				digest: params.ForkDigest(params.BeaconConfig().BellatrixForkEpoch),
				chain:  &mock.ChainService{ValidatorsRoot: [32]byte{}},
			},
			wantBlock: func() interfaces.ReadOnlySignedBeaconBlock {
				wsb, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockBellatrix{Block: &silapb.BeaconBlockBellatrix{Body: &silapb.BeaconBlockBodyBellatrix{SilaPayload: &silaenginev1.SilaPayload{}}}})
				require.NoError(t, err)
				return wsb
			}(),
			wantMd:          wrapper.WrappedMetadataV1(&silapb.MetaDataV1{}),
			wantAtt:         &silapb.Attestation{},
			wantAggregate:   &silapb.SignedAggregateAttestationAndProof{},
			wantAttSlashing: &silapb.AttesterSlashing{},
			wantErr:         false,
		},
		{
			name: "capella fork version",
			args: args{
				digest: params.ForkDigest(params.BeaconConfig().CapellaForkEpoch),
				chain:  &mock.ChainService{ValidatorsRoot: [32]byte{}},
			},
			wantBlock: func() interfaces.ReadOnlySignedBeaconBlock {
				wsb, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockCapella{Block: &silapb.BeaconBlockCapella{Body: &silapb.BeaconBlockBodyCapella{SilaPayload: &silaenginev1.SilaPayloadCapella{}}}})
				require.NoError(t, err)
				return wsb
			}(),
			wantMd:          wrapper.WrappedMetadataV1(&silapb.MetaDataV1{}),
			wantAtt:         &silapb.Attestation{},
			wantAggregate:   &silapb.SignedAggregateAttestationAndProof{},
			wantAttSlashing: &silapb.AttesterSlashing{},
			wantErr:         false,
		},
		{
			name: "deneb fork version",
			args: args{
				digest: params.ForkDigest(params.BeaconConfig().DenebForkEpoch),
				chain:  &mock.ChainService{ValidatorsRoot: [32]byte{}},
			},
			wantBlock: func() interfaces.ReadOnlySignedBeaconBlock {
				wsb, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockDeneb{Block: &silapb.BeaconBlockDeneb{Body: &silapb.BeaconBlockBodyDeneb{SilaPayload: &silaenginev1.SilaPayloadDeneb{}}}})
				require.NoError(t, err)
				return wsb
			}(),
			wantMd:          wrapper.WrappedMetadataV1(&silapb.MetaDataV1{}),
			wantAtt:         &silapb.Attestation{},
			wantAggregate:   &silapb.SignedAggregateAttestationAndProof{},
			wantAttSlashing: &silapb.AttesterSlashing{},
			wantErr:         false,
		},
		{
			name: "electra fork version",
			args: args{
				digest: params.ForkDigest(params.BeaconConfig().ElectraForkEpoch),
				chain:  &mock.ChainService{ValidatorsRoot: [32]byte{}},
			},
			wantBlock: func() interfaces.ReadOnlySignedBeaconBlock {
				wsb, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockElectra{Block: &silapb.BeaconBlockElectra{Body: &silapb.BeaconBlockBodyElectra{SilaPayload: &silaenginev1.SilaPayloadDeneb{}}}})
				require.NoError(t, err)
				return wsb
			}(),
			wantMd:          wrapper.WrappedMetadataV1(&silapb.MetaDataV1{}),
			wantAtt:         &silapb.SingleAttestation{},
			wantAggregate:   &silapb.SignedAggregateAttestationAndProofElectra{},
			wantAttSlashing: &silapb.AttesterSlashingElectra{},
			wantErr:         false,
		},
		{
			name: "fulu fork version",
			args: args{
				digest: params.ForkDigest(params.BeaconConfig().FuluForkEpoch),
				chain:  &mock.ChainService{ValidatorsRoot: [32]byte{}},
			},
			wantBlock: func() interfaces.ReadOnlySignedBeaconBlock {
				wsb, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockFulu{Block: &silapb.BeaconBlockElectra{Body: &silapb.BeaconBlockBodyElectra{SilaPayload: &silaenginev1.SilaPayloadDeneb{}}}})
				require.NoError(t, err)
				return wsb
			}(),
			wantMd:          wrapper.WrappedMetadataV1(&silapb.MetaDataV1{}),
			wantAtt:         &silapb.SingleAttestation{},
			wantAggregate:   &silapb.SignedAggregateAttestationAndProofElectra{},
			wantAttSlashing: &silapb.AttesterSlashingElectra{},
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBlock, err := extractDataTypeFromTypeMap(types.BlockMap, tt.args.digest[:], tt.args.chain)
			if (err != nil) != tt.wantErr {
				t.Errorf("block: error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotBlock, tt.wantBlock) {
				t.Errorf("block: got = %v, want %v", gotBlock, tt.wantBlock)
			}
			gotAtt, err := extractDataTypeFromTypeMap(types.AttestationMap, tt.args.digest[:], tt.args.chain)
			if (err != nil) != tt.wantErr {
				t.Errorf("attestation: error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotAtt, tt.wantAtt) {
				t.Errorf("attestation: got = %v, want %v", gotAtt, tt.wantAtt)
			}
			gotAggregate, err := extractDataTypeFromTypeMap(types.AggregateAttestationMap, tt.args.digest[:], tt.args.chain)
			if (err != nil) != tt.wantErr {
				t.Errorf("aggregate: error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotAggregate, tt.wantAggregate) {
				t.Errorf("aggregate: got = %v, want %v", gotAggregate, tt.wantAggregate)
			}
			gotAttSlashing, err := extractDataTypeFromTypeMap(types.AttesterSlashingMap, tt.args.digest[:], tt.args.chain)
			if (err != nil) != tt.wantErr {
				t.Errorf("attester slashing: error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotAttSlashing, tt.wantAttSlashing) {
				t.Errorf("attester slashin: got = %v, want %v", gotAttSlashing, tt.wantAttSlashing)
			}
		})
	}
}

func TestExtractDataTypeFromTypeMapInvalid(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	params.BeaconConfig().FuluForkEpoch = params.BeaconConfig().ElectraForkEpoch + 4096*2
	params.BeaconConfig().InitializeForkSchedule()
	chain := &mock.ChainService{ValidatorsRoot: [32]byte{}}
	_, err := extractDataTypeFromTypeMap(types.BlockMap, []byte{0x00, 0x01}, chain)
	require.ErrorIs(t, err, errInvalidDigest)
	_, err = extractDataTypeFromTypeMap(types.AttestationMap, []byte{0x00, 0x01}, chain)
	require.ErrorIs(t, err, errInvalidDigest)
}
