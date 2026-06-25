package p2p

import (
	"reflect"

	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"google.golang.org/protobuf/proto"
)

// gossipTopicMappings represent the protocol ID to protobuf message type map for easy
// lookup.
var gossipTopicMappings = map[string]func() proto.Message{
	BlockSubnetTopicFormat:                    func() proto.Message { return &silapb.SignedBeaconBlock{} },
	AttestationSubnetTopicFormat:              func() proto.Message { return &silapb.Attestation{} },
	ExitSubnetTopicFormat:                     func() proto.Message { return &silapb.SignedVoluntaryExit{} },
	ProposerSlashingSubnetTopicFormat:         func() proto.Message { return &silapb.ProposerSlashing{} },
	AttesterSlashingSubnetTopicFormat:         func() proto.Message { return &silapb.AttesterSlashing{} },
	AggregateAndProofSubnetTopicFormat:        func() proto.Message { return &silapb.SignedAggregateAttestationAndProof{} },
	SyncContributionAndProofSubnetTopicFormat: func() proto.Message { return &silapb.SignedContributionAndProof{} },
	SyncCommitteeSubnetTopicFormat:            func() proto.Message { return &silapb.SyncCommitteeMessage{} },
	BlsToSilaChangeSubnetTopicFormat:     func() proto.Message { return &silapb.SignedBLSToSilaChange{} },
	BlobSubnetTopicFormat:                     func() proto.Message { return &silapb.BlobSidecar{} },
	LightClientOptimisticUpdateTopicFormat:    func() proto.Message { return &silapb.LightClientOptimisticUpdateAltair{} },
	LightClientFinalityUpdateTopicFormat:      func() proto.Message { return &silapb.LightClientFinalityUpdateAltair{} },
	DataColumnSubnetTopicFormat:               func() proto.Message { return &silapb.DataColumnSidecar{} },
	PayloadAttestationMessageTopicFormat:      func() proto.Message { return &silapb.PayloadAttestationMessage{} },
	SilaPayloadEnvelopeTopicFormat:       func() proto.Message { return &silapb.SignedSilaPayloadEnvelope{} },
	SilaPayloadBidTopicFormat:            func() proto.Message { return &silapb.SignedSilaPayloadBid{} },
	SignedProposerPreferencesTopicFormat:      func() proto.Message { return &silapb.SignedProposerPreferences{} },
}

// GossipTopicMappings is a function to return the assigned data type
// versioned by epoch.
func GossipTopicMappings(topic string, epoch primitives.Epoch) proto.Message {
	switch topic {
	case BlockSubnetTopicFormat:
		if epoch >= params.BeaconConfig().FuluForkEpoch {
			return &silapb.SignedBeaconBlockFulu{}
		}
		if epoch >= params.BeaconConfig().GloasForkEpoch {
			return &silapb.SignedBeaconBlockGloas{}
		}
		if epoch >= params.BeaconConfig().ElectraForkEpoch {
			return &silapb.SignedBeaconBlockElectra{}
		}
		if epoch >= params.BeaconConfig().DenebForkEpoch {
			return &silapb.SignedBeaconBlockDeneb{}
		}
		if epoch >= params.BeaconConfig().CapellaForkEpoch {
			return &silapb.SignedBeaconBlockCapella{}
		}
		if epoch >= params.BeaconConfig().BellatrixForkEpoch {
			return &silapb.SignedBeaconBlockBellatrix{}
		}
		if epoch >= params.BeaconConfig().AltairForkEpoch {
			return &silapb.SignedBeaconBlockAltair{}
		}
		return gossipMessage(topic)
	case AttestationSubnetTopicFormat:
		if epoch >= params.BeaconConfig().ElectraForkEpoch {
			return &silapb.SingleAttestation{}
		}
		return gossipMessage(topic)
	case AttesterSlashingSubnetTopicFormat:
		if epoch >= params.BeaconConfig().ElectraForkEpoch {
			return &silapb.AttesterSlashingElectra{}
		}
		return gossipMessage(topic)
	case AggregateAndProofSubnetTopicFormat:
		if epoch >= params.BeaconConfig().ElectraForkEpoch {
			return &silapb.SignedAggregateAttestationAndProofElectra{}
		}
		return gossipMessage(topic)
	case LightClientOptimisticUpdateTopicFormat:
		if epoch >= params.BeaconConfig().DenebForkEpoch {
			return &silapb.LightClientOptimisticUpdateDeneb{}
		}
		if epoch >= params.BeaconConfig().CapellaForkEpoch {
			return &silapb.LightClientOptimisticUpdateCapella{}
		}
		return gossipMessage(topic)
	case LightClientFinalityUpdateTopicFormat:
		if epoch >= params.BeaconConfig().ElectraForkEpoch {
			return &silapb.LightClientFinalityUpdateElectra{}
		}
		if epoch >= params.BeaconConfig().DenebForkEpoch {
			return &silapb.LightClientFinalityUpdateDeneb{}
		}
		if epoch >= params.BeaconConfig().CapellaForkEpoch {
			return &silapb.LightClientFinalityUpdateCapella{}
		}
		return gossipMessage(topic)
	case DataColumnSubnetTopicFormat:
		if epoch >= params.BeaconConfig().GloasForkEpoch {
			return &silapb.DataColumnSidecarGloas{}
		}
		return gossipMessage(topic)
	default:
		return gossipMessage(topic)
	}
}

func gossipMessage(topic string) proto.Message {
	msgGen, ok := gossipTopicMappings[topic]
	if !ok {
		return nil
	}
	return msgGen()
}

// AllTopics returns all topics stored in our
// gossip mapping.
func AllTopics() []string {
	var topics []string
	for k := range gossipTopicMappings {
		topics = append(topics, k)
	}
	return topics
}

// GossipTypeMapping is the inverse of GossipTopicMappings so that an arbitrary protobuf message
// can be mapped to a protocol ID string.
var GossipTypeMapping = make(map[reflect.Type]string, len(gossipTopicMappings))

func init() {
	for k, v := range gossipTopicMappings {
		GossipTypeMapping[reflect.TypeOf(v())] = k
	}

	// Specially handle Altair objects.
	GossipTypeMapping[reflect.TypeFor[*silapb.SignedBeaconBlockAltair]()] = BlockSubnetTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.LightClientFinalityUpdateAltair]()] = LightClientFinalityUpdateTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.LightClientOptimisticUpdateAltair]()] = LightClientOptimisticUpdateTopicFormat

	// Specially handle Bellatrix objects.
	GossipTypeMapping[reflect.TypeFor[*silapb.SignedBeaconBlockBellatrix]()] = BlockSubnetTopicFormat

	// Specially handle Capella objects.
	GossipTypeMapping[reflect.TypeFor[*silapb.SignedBeaconBlockCapella]()] = BlockSubnetTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.LightClientOptimisticUpdateCapella]()] = LightClientOptimisticUpdateTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.LightClientFinalityUpdateCapella]()] = LightClientFinalityUpdateTopicFormat

	// Specially handle Deneb objects.
	GossipTypeMapping[reflect.TypeFor[*silapb.SignedBeaconBlockDeneb]()] = BlockSubnetTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.LightClientOptimisticUpdateDeneb]()] = LightClientOptimisticUpdateTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.LightClientFinalityUpdateDeneb]()] = LightClientFinalityUpdateTopicFormat

	// Specially handle Electra objects.
	GossipTypeMapping[reflect.TypeFor[*silapb.SignedBeaconBlockElectra]()] = BlockSubnetTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.SingleAttestation]()] = AttestationSubnetTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.AttesterSlashingElectra]()] = AttesterSlashingSubnetTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.SignedAggregateAttestationAndProofElectra]()] = AggregateAndProofSubnetTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.LightClientFinalityUpdateElectra]()] = LightClientFinalityUpdateTopicFormat

	// Specially handle Fulu objects.
	GossipTypeMapping[reflect.TypeFor[*silapb.SignedBeaconBlockFulu]()] = BlockSubnetTopicFormat
	// Specially handle Gloas objects.
	GossipTypeMapping[reflect.TypeFor[*silapb.SignedBeaconBlockGloas]()] = BlockSubnetTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.DataColumnSidecarGloas]()] = DataColumnSubnetTopicFormat

	// Payload attestation messages.
	GossipTypeMapping[reflect.TypeFor[*silapb.PayloadAttestationMessage]()] = PayloadAttestationMessageTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.SignedSilaPayloadBid]()] = SilaPayloadBidTopicFormat
	GossipTypeMapping[reflect.TypeFor[*silapb.SignedProposerPreferences]()] = SignedProposerPreferencesTopicFormat
}
